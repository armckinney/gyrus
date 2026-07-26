package git

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/armckinney/gyrus/internal/okf"
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
	repoURL     string
	branch      string
	repo        *git.Repository
	auth        transport.AuthMethod
	authorName  string
	authorEmail string
	fs          billy.Filesystem
}

type Options struct {
	RepoURL     string
	Branch      string
	AuthorName  string
	AuthorEmail string
}

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
		return nil, fmt.Errorf("failed to discover auth: %w", err)
	}

	fs := memfs.New()
	st := memory.NewStorage()

	store := &Store{
		repoURL:     opts.RepoURL,
		branch:      opts.Branch,
		auth:        auth,
		authorName:  opts.AuthorName,
		authorEmail: opts.AuthorEmail,
		fs:          fs,
	}

	repo, err := git.Clone(st, fs, &git.CloneOptions{
		URL:           opts.RepoURL,
		Auth:          auth,
		ReferenceName: plumbing.NewBranchReferenceName(opts.Branch),
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		if err == transport.ErrEmptyRemoteRepository || strings.Contains(err.Error(), "reference not found") || strings.Contains(err.Error(), "repository not found") {
			st = memory.NewStorage(); fs = memfs.New(); store.fs = fs; repo, err = git.Init(st, fs)
			if err != nil {
				return nil, fmt.Errorf("failed to init empty repo: %w", err)
			}
			_, err = repo.CreateRemote(&config.RemoteConfig{
				Name: "origin",
				URLs: []string{opts.RepoURL},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create remote: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to clone repo: %w", err)
		}
	}

	store.repo = repo
	return store, nil
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

func (s *Store) docPath(doc *gyrus.Document) string {
	ownerGroup := doc.OwnerGroup
	if ownerGroup == "" {
		ownerGroup = "default"
	}
	categorySubdir := "reference"
	if doc.Category == gyrus.CategoryBusinessLogic || doc.Category == gyrus.CategoryProduct {
		categorySubdir = "workspaces/main"
	} else if string(doc.Category) != "" {
		categorySubdir = string(doc.Category)
	}

	return filepath.Join("docs", "okf", ownerGroup, categorySubdir, fmt.Sprintf("%s.md", doc.ID))
}

func (s *Store) findDocPathByID(id string) (string, error) {
	targetMD := id
	if !strings.HasSuffix(targetMD, ".md") && !strings.HasSuffix(targetMD, ".json") {
		targetMD = id + ".md"
	}

	var found string
	var walk func(dir string) error
	walk = func(dir string) error {
		entries, err := s.fs.ReadDir(dir)
		if err != nil {
			return nil
		}
		for _, entry := range entries {
			relPath := filepath.Join(dir, entry.Name())
			if entry.IsDir() {
				if err := walk(relPath); err != nil {
					return err
				}
			} else if entry.Name() == targetMD || entry.Name() == id {
				found = relPath
				return nil
			}
		}
		return nil
	}

	_ = walk(".")
	if found != "" {
		return found, nil
	}

	return "", fmt.Errorf("document '%s' not found in git repository", id)
}

func (s *Store) Create(ctx context.Context, doc gyrus.Document) (gyrus.DocumentRef, error) {
	doc.Version = 1
	doc.LastUpdated = time.Now()

	data, err := okf.SerializeMarkdown(&doc)
	if err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to serialize doc: %w", err)
	}

	path := s.docPath(&doc)

	// Create directory if needed
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := s.fs.MkdirAll(dir, 0755); err != nil {
			return gyrus.DocumentRef{}, fmt.Errorf("failed to create dir: %w", err)
		}
	}

	f, err := s.fs.Create(path)
	if err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to create file: %w", err)
	}
	_, err = f.Write(data)
	f.Close()
	if err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to write file: %w", err)
	}

	wt, err := s.repo.Worktree()
	if err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = wt.Add(path)
	if err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to add file to index: %w", err)
	}

	msg := fmt.Sprintf("chore(docs): create document %s v%d", doc.ID, doc.Version)
	_, err = wt.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  s.authorName,
			Email: s.authorEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to commit: %w", err)
	}

	if err := s.push(); err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to push: %w", err)
	}

	return gyrus.DocumentRef{
		ID:          doc.ID,
		Version:     doc.Version,
		Status:      doc.Status,
		LastUpdated: doc.LastUpdated,
	}, nil
}

func (s *Store) Get(ctx context.Context, id string) (gyrus.Document, error) {
	path, err := s.findDocPathByID(id)
	if err != nil {
		return gyrus.Document{}, err
	}

	f, err := s.fs.Open(path)
	if err != nil {
		return gyrus.Document{}, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return gyrus.Document{}, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	if strings.HasSuffix(path, ".json") {
		doc, err := okf.ParseJSON(data)
		if err != nil {
			return gyrus.Document{}, err
		}
		return *doc, nil
	}

	doc, err := okf.ParseMarkdown(data)
	if err != nil {
		return gyrus.Document{}, err
	}
	return *doc, nil
}

func (s *Store) Update(ctx context.Context, id string, patch gyrus.DocumentPatch, expectedVersion int) (gyrus.DocumentRef, error) {
	doc, err := s.Get(ctx, id)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	if doc.Version != expectedVersion {
		return gyrus.DocumentRef{}, fmt.Errorf("version mismatch: expected %d, got %d", expectedVersion, doc.Version)
	}

	if patch.Title != nil {
		doc.Title = *patch.Title
	}
	if patch.Status != nil {
		doc.Status = *patch.Status
	}
	if patch.Tags != nil {
		doc.Tags = *patch.Tags
	}
	if patch.Dependencies != nil {
		doc.Dependencies = *patch.Dependencies
	}
	if patch.Content != nil {
		doc.Content = *patch.Content
	}

	doc.Version++
	doc.LastUpdated = time.Now()

	data, err := okf.SerializeMarkdown(&doc)
	if err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to serialize doc: %w", err)
	}

	path, err := s.findDocPathByID(id)
	if err != nil {
		path = s.docPath(&doc)
	}

	f, err := s.fs.Create(path)
	if err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to update file: %w", err)
	}
	_, err = f.Write(data)
	f.Close()
	if err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to write file: %w", err)
	}

	wt, err := s.repo.Worktree()
	if err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = wt.Add(path)
	if err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to add file to index: %w", err)
	}

	msg := fmt.Sprintf("chore(docs): update document %s to v%d", doc.ID, doc.Version)
	if patch.Reason != "" {
		msg += "\n\nReason: " + patch.Reason
	}
	_, err = wt.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  s.authorName,
			Email: s.authorEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to commit: %w", err)
	}

	if err := s.push(); err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to push: %w", err)
	}

	return gyrus.DocumentRef{
		ID:          doc.ID,
		Version:     doc.Version,
		Status:      doc.Status,
		LastUpdated: doc.LastUpdated,
	}, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	path, err := s.findDocPathByID(id)
	if err != nil {
		return err
	}
	
	wt, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = wt.Remove(path)
	if err != nil {
		return fmt.Errorf("failed to remove file from index: %w", err)
	}

	msg := fmt.Sprintf("chore(docs): delete document %s", id)
	_, err = wt.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  s.authorName,
			Email: s.authorEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return s.push()
}

func (s *Store) Archive(ctx context.Context, id string) error {
	doc, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	archivedStatus := "archived"
	patch := gyrus.DocumentPatch{
		Status: &archivedStatus,
		Reason: "archived",
	}
	_, err = s.Update(ctx, id, patch, doc.Version)
	return err
}

func (s *Store) push() error {
	err := s.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       s.auth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to push to remote git repository (%s): %w", s.repoURL, err)
	}
	return nil
}
