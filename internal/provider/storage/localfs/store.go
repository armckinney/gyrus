package localfs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// Store implements gyrus.ConceptStore over the local filesystem.
type Store struct {
	rootDir        string
	defaultProject string
	mu             sync.RWMutex
}

// NewStore initializes a new localfs ConceptStore at rootDir.
func NewStore(rootDir string) (*Store, error) {
	return NewStoreWithProject(rootDir, "")
}

// NewStoreWithProject initializes a new localfs ConceptStore at rootDir with default project.
func NewStoreWithProject(rootDir string, defaultProject string) (*Store, error) {
	absPath, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("invalid storage root path: %w", err)
	}
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage root directory: %w", err)
	}
	return &Store{
		rootDir:        absPath,
		defaultProject: defaultProject,
	}, nil
}

func parseScopeFromPath(filePath string) (scope string, project string) {
	clean := filepath.ToSlash(filePath)
	idxConcepts := strings.Index(clean, "/concepts/")
	if idxConcepts != -1 {
		sub := clean[idxConcepts+len("/concepts/"):] // "<tenant>/projects/<name>/..." or "<tenant>/reference/..."
		parts := strings.Split(sub, "/")
		if len(parts) >= 3 && parts[1] == "reference" {
			return "reference", ""
		}
		if len(parts) >= 3 && parts[1] == "projects" && parts[2] != "" {
			return "project", parts[2]
		}
	}
	if strings.Contains(clean, "/reference/") {
		return "reference", ""
	}
	if idx := strings.LastIndex(clean, "/projects/"); idx != -1 {
		rest := clean[idx+len("/projects/"):]
		parts := strings.Split(rest, "/")
		if len(parts) >= 1 && parts[0] != "" {
			return "project", parts[0]
		}
	}
	return "reference", ""
}

func (s *Store) getConceptPath(concept *gyrus.Concept) string {
	tenant := concept.Tenant
	if tenant == "" {
		tenant = "default"
	}

	var subdir string
	if concept.Project != "" {
		subdir = filepath.Join("projects", concept.Project)
	} else if concept.Scope == "reference" {
		subdir = "reference"
	} else if s.defaultProject != "" {
		subdir = filepath.Join("projects", s.defaultProject)
	} else if concept.Category == gyrus.CategoryBusinessLogic || concept.Category == gyrus.CategoryProduct {
		subdir = "projects/gyrus"
	} else {
		subdir = "reference"
	}

	return filepath.Join(s.rootDir, "concepts", tenant, subdir, fmt.Sprintf("%s.md", concept.ID))
}

func (s *Store) getConceptPathByID(id string) (string, error) {
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
		return "", fmt.Errorf("concept '%s' not found in storage root", id)
	}
	return found, nil
}

func (s *Store) Create(ctx context.Context, concept gyrus.Concept) (gyrus.ConceptRef, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := okf.Validate(&concept); err != nil {
		return gyrus.ConceptRef{}, fmt.Errorf("validation error: %w", err)
	}

	conceptPath := s.getConceptPath(&concept)
	if _, err := os.Stat(conceptPath); err == nil {
		return gyrus.ConceptRef{}, fmt.Errorf("concept '%s' already exists", concept.ID)
	}

	if err := os.MkdirAll(filepath.Dir(conceptPath), 0755); err != nil {
		return gyrus.ConceptRef{}, fmt.Errorf("failed to create parent directories: %w", err)
	}

	concept.LastUpdated = time.Now().Truncate(time.Second)
	payload, err := okf.SerializeMarkdown(&concept)
	if err != nil {
		return gyrus.ConceptRef{}, fmt.Errorf("failed to serialize concept: %w", err)
	}

	if err := os.WriteFile(conceptPath, payload, 0644); err != nil {
		return gyrus.ConceptRef{}, fmt.Errorf("failed to write concept file: %w", err)
	}

	return gyrus.ConceptRef{
		ID:          concept.ID,
		Version:     concept.Version,
		Status:      concept.Status,
		LastUpdated: concept.LastUpdated,
	}, nil
}

func (s *Store) Get(ctx context.Context, id string) (gyrus.Concept, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conceptPath, err := s.getConceptPathByID(id)
	if err != nil {
		return gyrus.Concept{}, err
	}

	data, err := os.ReadFile(conceptPath)
	if err != nil {
		return gyrus.Concept{}, fmt.Errorf("failed to read concept file: %w", err)
	}

	concept, err := okf.ParseMarkdown(data)
	if err != nil {
		return gyrus.Concept{}, fmt.Errorf("failed to parse concept file: %w", err)
	}

	scope, project := parseScopeFromPath(conceptPath)
	concept.FilePath = conceptPath
	concept.Scope = scope
	concept.Project = project

	return *concept, nil
}

func (s *Store) Update(ctx context.Context, id string, patch gyrus.ConceptPatch, expectedVersion int) (gyrus.ConceptRef, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conceptPath, err := s.getConceptPathByID(id)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}

	data, err := os.ReadFile(conceptPath)
	if err != nil {
		return gyrus.ConceptRef{}, fmt.Errorf("failed to read concept file: %w", err)
	}

	concept, err := okf.ParseMarkdown(data)
	if err != nil {
		return gyrus.ConceptRef{}, fmt.Errorf("failed to parse concept file: %w", err)
	}

	if expectedVersion > 0 && concept.Version != expectedVersion {
		return gyrus.ConceptRef{}, fmt.Errorf("concurrency error: expected version %d, but current version is %d", expectedVersion, concept.Version)
	}

	// Validate lifecycle transition and content immutability
	contentChanged := patch.Content != nil && *patch.Content != concept.Content
	if err := lifecycle.ValidateMutation(concept.Type, concept.Status, concept.Immutable, contentChanged); err != nil {
		return gyrus.ConceptRef{}, err
	}

	// Apply patch fields
	if patch.Title != nil {
		concept.Title = *patch.Title
	}
	if patch.Status != nil {
		if err := lifecycle.ValidateTransition(concept.Type, concept.Status, *patch.Status); err != nil {
			return gyrus.ConceptRef{}, err
		}
		concept.Status = *patch.Status
	}
	if patch.Tags != nil {
		concept.Tags = *patch.Tags
	}
	if patch.Dependencies != nil {
		concept.Dependencies = *patch.Dependencies
	}
	if patch.Content != nil {
		concept.Content = *patch.Content
	}

	concept.Version++
	concept.LastUpdated = time.Now().Truncate(time.Second)

	if err := okf.Validate(concept); err != nil {
		return gyrus.ConceptRef{}, fmt.Errorf("validation error after patch: %w", err)
	}

	payload, err := okf.SerializeMarkdown(concept)
	if err != nil {
		return gyrus.ConceptRef{}, fmt.Errorf("failed to serialize updated concept: %w", err)
	}

	if err := os.WriteFile(conceptPath, payload, 0644); err != nil {
		return gyrus.ConceptRef{}, fmt.Errorf("failed to write updated concept file: %w", err)
	}

	return gyrus.ConceptRef{
		ID:          concept.ID,
		Version:     concept.Version,
		Status:      concept.Status,
		LastUpdated: concept.LastUpdated,
	}, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conceptPath, err := s.getConceptPathByID(id)
	if err != nil {
		return err
	}

	return os.Remove(conceptPath)
}

func (s *Store) Archive(ctx context.Context, id string) error {
	// For localfs, archiving deletes the concept file from disk
	return s.Delete(ctx, id)
}

func (s *Store) schemasDir() string {
	if filepath.Base(s.rootDir) == "concepts" || filepath.Base(s.rootDir) == "docs" {
		return filepath.Join(filepath.Dir(s.rootDir), "schemas")
	}
	return filepath.Join(s.rootDir, "schemas")
}

func (s *Store) schemaPath(conceptType string) string {
	return filepath.Join(s.schemasDir(), fmt.Sprintf("%s.md", conceptType))
}

// GetSchema retrieves a persisted schema template for a concept type.
func (s *Store) GetSchema(ctx context.Context, conceptType string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	targetPath := s.schemaPath(conceptType)
	if err := okf.ValidateSchemaPath(filepath.ToSlash(targetPath)); err != nil {
		return "", err
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveSchema persists an OKF schema template to the enforced .gyrus/schemas/ path.
func (s *Store) SaveSchema(ctx context.Context, conceptType string, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := okf.ValidateSchemaContent(conceptType, content); err != nil {
		return err
	}

	dir := s.schemasDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create schemas directory: %w", err)
	}

	targetPath := s.schemaPath(conceptType)
	if err := okf.ValidateSchemaPath(filepath.ToSlash(targetPath)); err != nil {
		return err
	}

	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed writing schema file: %w", err)
	}

	return nil
}

// ListSchemas enumerates all persisted schema template types under .gyrus/schemas/.
func (s *Store) ListSchemas(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := s.schemasDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var list []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			name := strings.TrimSuffix(e.Name(), ".md")
			list = append(list, name)
		}
	}
	sort.Strings(list)
	return list, nil
}

// DeleteSchema removes a persisted schema template from .gyrus/schemas/.
func (s *Store) DeleteSchema(ctx context.Context, conceptType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	targetPath := s.schemaPath(conceptType)
	if err := okf.ValidateSchemaPath(filepath.ToSlash(targetPath)); err != nil {
		return err
	}

	if err := os.Remove(targetPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("schema '%s' not found: %w", conceptType, err)
		}
		return err
	}
	return nil
}
