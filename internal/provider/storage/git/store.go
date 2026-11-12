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
	repoURL, branch, trunkBranch, strategy, authorName, authorEmail string
	repo                                                            *git.Repository
	auth                                                            transport.AuthMethod
	fs                                                              billy.Filesystem
	mu                                                              sync.Mutex
}

type Options struct {
	RepoURL, TrunkBranch, CommitStrategy, AuthorName, AuthorEmail string
	// Branch is a runtime initiative selection; it is never persisted in config.
	Branch string
}

type Status struct {
	Strategy          string `json:"strategy"`
	TrunkBranch       string `json:"trunk_branch"`
	ActiveBranch      string `json:"active_branch"`
	SelectionRequired bool   `json:"selection_required"`
}

func NewStore(opts Options) (*Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return NewStoreContext(ctx, opts)
}

func NewStoreContext(ctx context.Context, opts Options) (*Store, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if opts.TrunkBranch == "" {
		opts.TrunkBranch = "main"
	}
	if opts.CommitStrategy == "" {
		opts.CommitStrategy = "direct"
	}
	if opts.AuthorName == "" {
		opts.AuthorName = "Gyrus Bot"
	}
	if opts.AuthorEmail == "" {
		opts.AuthorEmail = "gyrus@localhost"
	}
	if opts.RepoURL == "" || strings.ContainsAny(opts.RepoURL, "\r\n") {
		return nil, fmt.Errorf("Git repo URL is required and must not contain newlines")
	}
	endpoint, err := transport.NewEndpoint(opts.RepoURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Git repo URL")
	}
	if endpoint.Password != "" {
		return nil, fmt.Errorf("Git repo URL must not contain passwords; use a credential helper")
	}
	s := &Store{repoURL: opts.RepoURL, trunkBranch: opts.TrunkBranch, authorName: opts.AuthorName, authorEmail: opts.AuthorEmail}
	if err := plumbing.NewBranchReferenceName(s.trunkBranch).Validate(); err != nil {
		return nil, fmt.Errorf("invalid trunk branch: %w", err)
	}
	if err := s.selectBranch(opts.CommitStrategy, opts.Branch); err != nil {
		return nil, err
	}
	// A branching MCP session may start before an initiative is selected.
	if s.branch != "" {
		if err := s.refresh(ctx); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Store) selectBranch(strategy, branch string) error {
	switch strategy {
	case "direct":
		if branch != "" {
			return fmt.Errorf("direct strategy does not accept an initiative branch")
		}
		branch = s.trunkBranch
	case "branching":
		if branch != "" {
			if err := plumbing.NewBranchReferenceName(branch).Validate(); err != nil {
				return fmt.Errorf("invalid initiative branch: %w", err)
			}
			if branch == s.trunkBranch {
				return fmt.Errorf("initiative branch must differ from trunk")
			}
		}
	default:
		return fmt.Errorf("invalid commit strategy: choose direct or branching")
	}
	s.strategy, s.branch = strategy, branch
	return nil
}

func (s *Store) Status() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return Status{Strategy: s.strategy, TrunkBranch: s.trunkBranch, ActiveBranch: s.branch, SelectionRequired: s.branch == ""}
}

// Switch prepares a new selection before replacing this session's active branch.
// It never changes remote refs or configuration files.
func (s *Store) Switch(ctx context.Context, strategy, branch string) (Status, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	candidate := &Store{repoURL: s.repoURL, trunkBranch: s.trunkBranch, authorName: s.authorName, authorEmail: s.authorEmail}
	if strategy == "branching" && branch == "" {
		return Status{}, fmt.Errorf("branching requires an initiative branch")
	}
	if err := candidate.selectBranch(strategy, branch); err != nil {
		return Status{}, err
	}
	if err := candidate.refresh(ctx); err != nil {
		return Status{}, err
	}
	s.strategy, s.branch, s.repo, s.fs, s.auth = candidate.strategy, candidate.branch, candidate.repo, candidate.fs, candidate.auth
	return Status{Strategy: s.strategy, TrunkBranch: s.trunkBranch, ActiveBranch: s.branch}, nil
}

// Every operation reloads confirmed remote state. Missing initiatives are based
// on trunk locally, and only become remote branches on the first mutation.
func (s *Store) refresh(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.branch == "" {
		return fmt.Errorf("branching strategy requires an initiative branch: use --git-branch or gyrus_git_switch")
	}
	auth, err := discoverAuth(ctx, s.repoURL)
	if err != nil {
		return err
	}
	s.auth = auth
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{Name: "origin", URLs: []string{s.repoURL}})
	refs, err := remote.ListContext(ctx, &git.ListOptions{Auth: s.auth})
	if err != nil && !errors.Is(err, transport.ErrEmptyRemoteRepository) {
		return s.remoteError("list", err)
	}
	exists := map[plumbing.ReferenceName]bool{}
	for _, ref := range refs {
		if ref.Name().IsBranch() {
			exists[ref.Name()] = true
		}
	}
	trunk, target := plumbing.NewBranchReferenceName(s.trunkBranch), plumbing.NewBranchReferenceName(s.branch)
	fs, st := memfs.New(), memory.NewStorage()
	var repo *git.Repository
	if len(exists) == 0 && s.strategy == "direct" {
		repo, err = git.InitWithOptions(st, fs, git.InitOptions{DefaultBranch: trunk})
		if err == nil {
			_, err = repo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{s.repoURL}})
		}
	} else {
		if !exists[trunk] {
			return fmt.Errorf("Git trunk branch %q does not exist", s.trunkBranch)
		}
		source := target
		if !exists[target] {
			source = trunk
		}
		repo, err = git.CloneContext(ctx, st, fs, &git.CloneOptions{URL: s.repoURL, Auth: s.auth, ReferenceName: source, SingleBranch: true})
		if err == nil && source != target {
			var wt *git.Worktree
			wt, err = repo.Worktree()
			if err == nil {
				err = wt.Checkout(&git.CheckoutOptions{Branch: target, Create: true})
			}
		}
	}
	if err != nil {
		return s.remoteError("refresh", err)
	}
	s.repo, s.fs = repo, fs
	return nil
}
func (s *Store) remoteError(op string, err error) error {
	message := strings.ReplaceAll(err.Error(), s.repoURL, "remote")
	if auth, ok := s.auth.(*http.BasicAuth); ok && auth.Password != "" {
		message = strings.ReplaceAll(message, auth.Password, "[redacted]")
	}
	return fmt.Errorf("Git %s failed: %s: %w", op, message, remoteCause(err))
}

// Preserve cancellation identity without retaining transport errors containing secrets.
func remoteCause(err error) error {
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}
	return errRemote
}

var errRemote = errors.New("remote operation failed")

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
func (s *Store) commit(ctx context.Context, id, message, name string, remove bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
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
	status, err := wt.Status()
	if err != nil {
		return err
	}
	if status.IsClean() {
		return nil
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
	status := "archived"
	_, err := s.Update(ctx, id, gyrus.ConceptPatch{Status: &status, Reason: "archived"}, 0)
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

func discoverAuth(ctx context.Context, repoURL string) (transport.AuthMethod, error) {
	endpoint, err := transport.NewEndpoint(repoURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Git repo URL")
	}
	if endpoint.Protocol == "http" || endpoint.Protocol == "https" {
		return resolveGitCredentialHelper(ctx, repoURL)
	}
	if endpoint.Protocol == "ssh" {
		user := endpoint.User
		if user == "" {
			user = "git"
		}
		if agent, err := ssh.NewSSHAgentAuth(user); err == nil {
			return agent, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		for _, keyName := range []string{"id_ed25519", "id_rsa", "id_ecdsa"} {
			if key, err := ssh.NewPublicKeysFromFile(user, filepath.Join(home, ".ssh", keyName), ""); err == nil {
				return key, nil
			}
		}
	}
	return nil, nil
}

func resolveGitCredentialHelper(ctx context.Context, repoURL string) (transport.AuthMethod, error) {
	if strings.ContainsAny(repoURL, "\r\n") {
		return nil, fmt.Errorf("Git repo URL must not contain newlines")
	}
	cmd := exec.CommandContext(ctx, "git", "credential", "fill")
	cmd.WaitDelay = time.Second
	// Disable terminal and GUI prompts, including Git Credential Manager prompts.
	for _, env := range os.Environ() {
		key, _, _ := strings.Cut(env, "=")
		if key != "GIT_TERMINAL_PROMPT" && key != "GCM_INTERACTIVE" && key != "GIT_ASKPASS" && key != "SSH_ASKPASS" {
			cmd.Env = append(cmd.Env, env)
		}
	}
	cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=0", "GCM_INTERACTIVE=never", "GIT_ASKPASS=", "SSH_ASKPASS=")
	cmd.Stdin = strings.NewReader("url=" + repoURL + "\n\n")
	out, err := cmd.Output()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		// No stored credentials is normal for public HTTPS repositories. Attempt
		// anonymous transport; private remotes then report an authentication error.
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return nil, nil
		}
		return nil, fmt.Errorf("Git credential helper unavailable; install Git and configure a credential manager")
	}
	var username, password string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), "=")
		if ok {
			switch key {
			case "username":
				username = value
			case "password":
				password = value
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("invalid credential helper response")
	}
	if username != "" && password != "" {
		return &http.BasicAuth{Username: username, Password: password}, nil
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
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	return s.commit(ctx, "", "chore(schemas): save "+docType, path, false)
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

	return s.commit(ctx, "", "chore(schemas): delete "+docType, path, true)
}
