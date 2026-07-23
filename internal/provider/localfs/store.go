package localfs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/armckinney/gyrus/internal/lifecycle"
	"github.com/armckinney/gyrus/internal/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// Store implements gyrus.DocumentStore over the local filesystem.
type Store struct {
	rootDir string
	mu      sync.RWMutex
}

// NewStore initializes a new localfs DocumentStore at rootDir.
func NewStore(rootDir string) (*Store, error) {
	absPath, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("invalid storage root path: %w", err)
	}
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage root directory: %w", err)
	}
	return &Store{rootDir: absPath}, nil
}

func (s *Store) getDocPath(doc *gyrus.Document) string {
	ownerGroup := doc.OwnerGroup
	if ownerGroup == "" {
		ownerGroup = "default"
	}

	categorySubdir := "reference"
	if doc.Category == gyrus.CategoryBusinessLogic || doc.Category == gyrus.CategoryProduct {
		categorySubdir = "workspaces/main"
	}

	return filepath.Join(s.rootDir, "okf", ownerGroup, categorySubdir, fmt.Sprintf("%s.md", doc.ID))
}

func (s *Store) getDocPathByID(id string) (string, error) {
	var found string
	err := filepath.Walk(s.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Name() == fmt.Sprintf("%s.md", id) {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("document '%s' not found in storage root", id)
	}
	return found, nil
}

func (s *Store) Create(ctx context.Context, doc gyrus.Document) (gyrus.DocumentRef, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := okf.Validate(&doc); err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("validation error: %w", err)
	}

	docPath := s.getDocPath(&doc)
	if _, err := os.Stat(docPath); err == nil {
		return gyrus.DocumentRef{}, fmt.Errorf("document '%s' already exists", doc.ID)
	}

	if err := os.MkdirAll(filepath.Dir(docPath), 0755); err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to create parent directories: %w", err)
	}

	doc.LastUpdated = time.Now().Truncate(time.Second)
	payload, err := okf.SerializeMarkdown(&doc)
	if err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to serialize document: %w", err)
	}

	if err := os.WriteFile(docPath, payload, 0644); err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to write document file: %w", err)
	}

	return gyrus.DocumentRef{
		ID:          doc.ID,
		Version:     doc.Version,
		Status:      doc.Status,
		LastUpdated: doc.LastUpdated,
	}, nil
}

func (s *Store) Get(ctx context.Context, id string) (gyrus.Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	docPath, err := s.getDocPathByID(id)
	if err != nil {
		return gyrus.Document{}, err
	}

	data, err := os.ReadFile(docPath)
	if err != nil {
		return gyrus.Document{}, fmt.Errorf("failed to read document file: %w", err)
	}

	doc, err := okf.ParseMarkdown(data)
	if err != nil {
		return gyrus.Document{}, fmt.Errorf("failed to parse document file: %w", err)
	}

	return *doc, nil
}

func (s *Store) Update(ctx context.Context, id string, patch gyrus.DocumentPatch, expectedVersion int) (gyrus.DocumentRef, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	docPath, err := s.getDocPathByID(id)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	data, err := os.ReadFile(docPath)
	if err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to read document file: %w", err)
	}

	doc, err := okf.ParseMarkdown(data)
	if err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to parse document file: %w", err)
	}

	if expectedVersion > 0 && doc.Version != expectedVersion {
		return gyrus.DocumentRef{}, fmt.Errorf("concurrency error: expected version %d, but current version is %d", expectedVersion, doc.Version)
	}

	// Validate lifecycle transition and content immutability
	contentChanged := patch.Content != nil && *patch.Content != doc.Content
	if err := lifecycle.ValidateMutation(doc.Type, doc.Status, doc.Immutable, contentChanged); err != nil {
		return gyrus.DocumentRef{}, err
	}

	// Apply patch fields
	if patch.Title != nil {
		doc.Title = *patch.Title
	}
	if patch.Status != nil {
		if err := lifecycle.ValidateTransition(doc.Type, doc.Status, *patch.Status); err != nil {
			return gyrus.DocumentRef{}, err
		}
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
	doc.LastUpdated = time.Now().Truncate(time.Second)

	if err := okf.Validate(doc); err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("validation error after patch: %w", err)
	}

	payload, err := okf.SerializeMarkdown(doc)
	if err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to serialize updated document: %w", err)
	}

	if err := os.WriteFile(docPath, payload, 0644); err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("failed to write updated document file: %w", err)
	}

	return gyrus.DocumentRef{
		ID:          doc.ID,
		Version:     doc.Version,
		Status:      doc.Status,
		LastUpdated: doc.LastUpdated,
	}, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	docPath, err := s.getDocPathByID(id)
	if err != nil {
		return err
	}

	return os.Remove(docPath)
}

func (s *Store) Archive(ctx context.Context, id string) error {
	// For localfs, archiving deletes the document file from disk
	return s.Delete(ctx, id)
}
