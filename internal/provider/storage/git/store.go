package git

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
)

type Store struct {
	repoURL, branch, authorName, authorEmail string
	repo                                     *git.Repository
	auth                                     transport.AuthMethod
	fs                                       billy.Filesystem
	mu                                       sync.Mutex
}
type Options struct{ RepoURL, Branch, AuthorName, AuthorEmail string }

func NewStore(opts Options) (*Store, error) {
	if opts.Branch == "" {
		opts.Branch = "main"
	}
	if opts.AuthorName == "" {
		opts.AuthorName = "Gyrus Bot"
	}
	if opts.AuthorEmail == "" {
		opts.AuthorEmail = "gyrus@localhost"
	}
	auth, err := discoverAuth(opts.RepoURL)
	if err != nil {
		return nil, fmt.Errorf("discover Git authentication: %w", err)
	}
	s := &Store{repoURL: opts.RepoURL, branch: opts.Branch, authorName: opts.AuthorName, authorEmail: opts.AuthorEmail, auth: auth}
	if err := s.refresh(context.Background()); err != nil {
		return nil, err
	}
	return s, nil
}

// Refresh by replacing the clone so a failed push can never leak unconfirmed data.
func (s *Store) refresh(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	fs, st := memfs.New(), memory.NewStorage()
	repo, err := git.CloneContext(ctx, st, fs, &git.CloneOptions{URL: s.repoURL, Auth: s.auth, ReferenceName: plumbing.NewBranchReferenceName(s.branch), SingleBranch: true})
	if errors.Is(err, transport.ErrEmptyRemoteRepository) {
		fs, st = memfs.New(), memory.NewStorage()
		repo, err = git.InitWithOptions(st, fs, git.InitOptions{DefaultBranch: plumbing.NewBranchReferenceName(s.branch)})
		if err == nil {
			_, err = repo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{s.repoURL}})
		}
	}
	if err != nil {
		return s.remoteError("refresh", err)
	}
	s.repo, s.fs = repo, fs
	return nil
}
func (s *Store) remoteError(op string, err error) error {
	return fmt.Errorf("Git %s failed: %s", op, strings.ReplaceAll(err.Error(), s.repoURL, "remote"))
}
func (s *Store) pushContext(ctx context.Context, id string) error {
	err := s.repo.PushContext(ctx, &git.PushOptions{RemoteName: "origin", Auth: s.auth, RefSpecs: []config.RefSpec{config.RefSpec("refs/heads/" + s.branch + ":refs/heads/" + s.branch)}})
	if err == nil || errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	}
	s.repo, s.fs = nil, nil
	if strings.Contains(err.Error(), "non-fast-forward") || strings.Contains(err.Error(), "failed to update ref") || strings.Contains(err.Error(), "reference has changed") {
		return &gyrus.ConflictError{ConceptID: id, Operation: "Git push", Err: s.remoteError("push", err)}
	}
	return s.remoteError("push", err)
}
func (s *Store) push() error { return s.pushContext(context.Background(), "") }
func (s *Store) commit(ctx context.Context, id, message, name string, remove bool) error {
	wt, err := s.repo.Worktree()
	if err != nil {
		return err
	}
	if remove {
		_, err = wt.Remove(name)
	} else {
		_, err = wt.Add(name)
	}
	if err != nil {
		return err
	}
	_, err = wt.Commit(message, &git.CommitOptions{Author: &object.Signature{Name: s.authorName, Email: s.authorEmail, When: time.Now().UTC()}})
	if err != nil {
		return err
	}
	return s.pushContext(ctx, id)
}
func (s *Store) paths() ([]string, error) {
	var result []string
	var walk func(string) error
	walk = func(dir string) error {
		entries, err := s.fs.ReadDir(dir)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		for _, entry := range entries {
			name := filepath.ToSlash(filepath.Join(dir, entry.Name()))
			if entry.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("concept collection contains symlink: %s", name)
			}
			if entry.IsDir() {
				if err := walk(name); err != nil {
					return err
				}
			} else if !strings.HasPrefix(entry.Name(), ".") && (strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".json")) {
				result = append(result, name)
			}
		}
		return nil
	}
	err := walk(".gyrus/concepts")
	sort.Strings(result)
	return result, err
}
func (s *Store) findDocPathByID(id string) (string, error) {
	if err := okf.ValidateName(id); err != nil {
		return "", err
	}
	found := ""
	for _, ext := range []string{".md", ".json"} {
		name := ".gyrus/concepts/" + id + ext
		if _, err := s.fs.Stat(name); err == nil {
			if found != "" {
				return "", fmt.Errorf("duplicate concept name %q", id)
			}
			found = name
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}
	if found == "" {
		return "", fmt.Errorf("concept %q: %w", id, os.ErrNotExist)
	}
	return found, nil
}
func (s *Store) read(name string) (gyrus.Concept, error) {
	f, err := s.fs.Open(name)
	if err != nil {
		return gyrus.Concept{}, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return gyrus.Concept{}, err
	}
	var doc *gyrus.Concept
	if strings.HasSuffix(name, ".json") {
		doc, err = okf.ParseJSON(data)
	} else {
		doc, err = okf.ParseMarkdown(data)
	}
	if err != nil {
		return gyrus.Concept{}, err
	}
	if err := okf.SetLocation(doc, name); err != nil {
		return gyrus.Concept{}, err
	}
	if doc.ID != strings.TrimSuffix(filepath.Base(name), filepath.Ext(name)) {
		return gyrus.Concept{}, fmt.Errorf("concept ID does not match filename: %s", name)
	}
	if err := okf.Validate(doc); err != nil {
		return gyrus.Concept{}, err
	}
	return *doc, nil
}
func (s *Store) get(id string) (gyrus.Concept, error) {
	name, err := s.findDocPathByID(id)
	if err != nil {
		return gyrus.Concept{}, err
	}
	return s.read(name)
}
func (s *Store) write(name string, doc gyrus.Concept) error {
	var data []byte
	var err error
	if strings.HasSuffix(name, ".json") {
		data, err = jsonFormat(&doc)
	} else {
		data, err = okf.SerializeMarkdown(&doc)
	}
	if err != nil {
		return err
	}
	if err := s.fs.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return err
	}
	f, err := s.fs.Create(name)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func jsonFormat(doc *gyrus.Concept) ([]byte, error) { return okf.SerializeMetadata(*doc) }
func conceptRef(doc gyrus.Concept) gyrus.ConceptRef {
	return gyrus.ConceptRef{Name: doc.Name, ID: doc.ID, Version: doc.Version, Status: doc.Status, LastUpdated: doc.LastUpdated}
}
func (s *Store) Create(ctx context.Context, doc gyrus.Concept) (gyrus.ConceptRef, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(ctx); err != nil {
		return gyrus.ConceptRef{}, err
	}
	doc.Version = 1
	doc.LastUpdated = time.Now().UTC().Truncate(time.Second)
	if err := okf.SetName(&doc, ""); err != nil {
		return gyrus.ConceptRef{}, err
	}
	if err := okf.Validate(&doc); err != nil {
		return gyrus.ConceptRef{}, err
	}
	if _, err := s.findDocPathByID(doc.Name); err == nil {
		return gyrus.ConceptRef{}, &gyrus.ConflictError{ConceptID: doc.Name, Operation: "create", Err: os.ErrExist}
	} else if !errors.Is(err, os.ErrNotExist) {
		return gyrus.ConceptRef{}, err
	}
	relative, err := okf.ConceptPath(doc, "")
	if err != nil {
		return gyrus.ConceptRef{}, err
	}
	name := ".gyrus/" + relative
	if err := s.write(name, doc); err != nil {
		return gyrus.ConceptRef{}, err
	}
	if err := s.commit(ctx, doc.Name, "chore(concepts): create "+doc.ID, name, false); err != nil {
		return gyrus.ConceptRef{}, err
	}
	return conceptRef(doc), nil
}
func (s *Store) Get(ctx context.Context, id string) (gyrus.Concept, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := okf.ValidateName(id); err != nil {
		return gyrus.Concept{}, err
	}
	if err := s.refresh(ctx); err != nil {
		return gyrus.Concept{}, err
	}
	return s.get(id)
}
func (s *Store) Update(ctx context.Context, id string, patch gyrus.ConceptPatch, version int) (gyrus.ConceptRef, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := okf.ValidateName(id); err != nil {
		return gyrus.ConceptRef{}, err
	}
	if err := s.refresh(ctx); err != nil {
		return gyrus.ConceptRef{}, err
	}
	doc, err := s.get(id)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}
	if version > 0 && version != doc.Version {
		return gyrus.ConceptRef{}, &gyrus.ConflictError{ConceptID: id, Operation: "update", Err: fmt.Errorf("expected version %d, got %d", version, doc.Version)}
	}
	if err := lifecycle.ApplyPatch(&doc, patch); err != nil {
		return gyrus.ConceptRef{}, err
	}
	if err := s.write(doc.FilePath, doc); err != nil {
		return gyrus.ConceptRef{}, err
	}
	message := "chore(concepts): update " + id
	if patch.Reason != "" {
		message += "\n\nReason: " + patch.Reason
	}
	if err := s.commit(ctx, id, message, doc.FilePath, false); err != nil {
		return gyrus.ConceptRef{}, err
	}
	return conceptRef(doc), nil
}
func (s *Store) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := okf.ValidateName(id); err != nil {
		return err
	}
	if err := s.refresh(ctx); err != nil {
		return err
	}
	name, err := s.findDocPathByID(id)
	if err != nil {
		return err
	}
	return s.commit(ctx, id, "chore(concepts): delete "+id, name, true)
}
func (s *Store) Archive(ctx context.Context, id string) error {
	doc, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	status := "archived"
	_, err = s.Update(ctx, id, gyrus.ConceptPatch{Status: &status, Reason: "archived"}, doc.Version)
	return err
}
func (s *Store) Scan(ctx context.Context, visit func(gyrus.Concept) error) (gyrus.SyncReport, error) {
	docs, report, err := s.snapshot(ctx)
	if err != nil {
		return report, err
	}
	return okf.Visit(ctx, docs, report, visit)
}
func (s *Store) snapshot(ctx context.Context) ([]gyrus.Concept, gyrus.SyncReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var report gyrus.SyncReport
	if err := s.refresh(ctx); err != nil {
		return nil, report, err
	}
	paths, err := s.paths()
	if err != nil {
		return nil, report, err
	}
	var docs []gyrus.Concept
	for _, name := range paths {
		if err := ctx.Err(); err != nil {
			return nil, report, err
		}
		report.ScannedFiles++
		doc, err := s.read(name)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		docs = append(docs, doc)
	}
	return docs, report, nil
}

func discoverAuth(url string) (transport.AuthMethod, error) {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		// 1. Check explicit environment variables first
		token := os.Getenv("GIT_AUTH_TOKEN")
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}
		if token != "" {
			return &http.BasicAuth{
				Username: "git",
				Password: token,
			}, nil
		}

		// 2. Query system git credential helper (git credential fill)
		if auth, err := resolveGitCredentialHelper(url); err == nil && auth != nil {
			return auth, nil
		}

		return nil, nil
	}

	if strings.HasPrefix(url, "git@") || strings.HasPrefix(url, "ssh://") {
		// 1. Try running SSH Agent auth if available
		if sshAgent, err := ssh.NewSSHAgentAuth("git"); err == nil && sshAgent != nil {
			return sshAgent, nil
		}

		// 2. Fall back to scanning user SSH key files
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		for _, keyName := range []string{"id_ed25519", "id_rsa", "id_ecdsa"} {
			keyPath := filepath.Join(home, ".ssh", keyName)
			if _, err := os.Stat(keyPath); err == nil {
				publicKeys, err := ssh.NewPublicKeysFromFile("git", keyPath, "")
				if err == nil {
					return publicKeys, nil
				}
			}
		}
		return nil, nil
	}
	return nil, nil
}

func resolveGitCredentialHelper(repoURL string) (transport.AuthMethod, error) {
	cmd := exec.Command("git", "credential", "fill")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("url=%s\n\n", repoURL))
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var username, password string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			switch parts[0] {
			case "username":
				username = parts[1]
			case "password":
				password = parts[1]
			}
		}
	}

	if password != "" {
		if username == "" {
			username = "git"
		}
		return &http.BasicAuth{
			Username: username,
			Password: password,
		}, nil
	}

	return nil, nil
}

func (s *Store) schemaPath(docType string) string {
	return filepath.ToSlash(filepath.Join(".gyrus", "schemas", fmt.Sprintf("%s.md", docType)))
}

// GetSchema retrieves a persisted schema template for a document type from Git storage.
func (s *Store) GetSchema(ctx context.Context, docType string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(ctx); err != nil {
		return "", err
	}
	path := s.schemaPath(docType)
	if err := okf.ValidateSchemaPath(path); err != nil {
		return "", err
	}

	f, err := s.fs.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveSchema persists an OKF schema template to .gyrus/schemas/ in Git storage.
func (s *Store) SaveSchema(ctx context.Context, docType string, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(ctx); err != nil {
		return err
	}
	if err := okf.ValidateSchemaContent(docType, content); err != nil {
		return err
	}

	path := s.schemaPath(docType)
	if err := okf.ValidateSchemaPath(path); err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := s.fs.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create schemas dir: %w", err)
	}

	f, err := s.fs.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create schema file: %w", err)
	}
	_, err = f.Write([]byte(content))
	f.Close()
	if err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	wt, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	if _, err := wt.Add(path); err != nil {
		return fmt.Errorf("failed to add schema file to git: %w", err)
	}

	msg := fmt.Sprintf("chore(schemas): save %s schema template", docType)
	if _, err := wt.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  s.authorName,
			Email: s.authorEmail,
			When:  time.Now(),
		},
	}); err != nil {
		return fmt.Errorf("failed to commit schema: %w", err)
	}

	return s.push()
}

// ListSchemas enumerates all persisted schema template types under .gyrus/schemas/ in Git storage.
func (s *Store) ListSchemas(ctx context.Context) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(ctx); err != nil {
		return nil, err
	}
	infos, err := s.fs.ReadDir(".gyrus/schemas")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var list []string
	for _, info := range infos {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			list = append(list, strings.TrimSuffix(info.Name(), ".md"))
		}
	}
	sort.Strings(list)
	return list, nil
}

// DeleteSchema removes a persisted schema template from Git storage.
func (s *Store) DeleteSchema(ctx context.Context, docType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(ctx); err != nil {
		return err
	}
	path := s.schemaPath(docType)
	if err := okf.ValidateSchemaPath(path); err != nil {
		return err
	}

	if _, err := s.fs.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("schema '%s' not found: %w", docType, err)
		}
		return err
	}

	if err := s.fs.Remove(path); err != nil {
		return fmt.Errorf("failed to remove schema file: %w", err)
	}

	wt, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	_, _ = wt.Remove(path)

	msg := fmt.Sprintf("chore(schemas): delete %s schema template", docType)
	if _, err := wt.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  s.authorName,
			Email: s.authorEmail,
			When:  time.Now(),
		},
	}); err != nil {
		return fmt.Errorf("failed to commit schema deletion: %w", err)
	}

	return s.push()
}
