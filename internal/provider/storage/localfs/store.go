package localfs

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/internal/provider/storage/fileio"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// Store implements concept and schema persistence on a local filesystem.
type Store struct {
	rootDir, defaultProject string
	mu                      sync.RWMutex
}

func NewStore(root string) (*Store, error) { return NewStoreWithProject(root, "") }
func NewStoreWithProject(root, project string) (*Store, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if project != "" {
		if err := okf.ValidateSegment("project", project); err != nil {
			return nil, err
		}
	}
	if err := os.MkdirAll(abs, fileio.DirectoryMode); err != nil {
		return nil, err
	}
	abs, err = filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, err
	}
	return &Store{rootDir: abs, defaultProject: project}, nil
}

func (s *Store) paths() ([]string, error) {
	var paths []string
	err := filepath.WalkDir(filepath.Join(s.rootDir, "concepts"), func(name string, e fs.DirEntry, err error) error {
		if errors.Is(err, os.ErrNotExist) && name == filepath.Join(s.rootDir, "concepts") {
			return nil
		}
		if err != nil {
			return err
		}
		if e.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("concept collection contains symlink: %s", name)
		}
		if !e.IsDir() && strings.HasSuffix(name, ".md") && !strings.HasPrefix(e.Name(), ".") {
			paths = append(paths, name)
		}
		return nil
	})
	return paths, err
}

func (s *Store) getConceptPathByID(id string) (string, error) {
	if err := okf.ValidateName(id); err != nil {
		return "", err
	}
	name, err := fileio.SafePath(s.rootDir, "concepts/"+id+".md")
	if err != nil {
		return "", err
	}
	if _, err = os.Stat(name); err != nil {
		return "", err
	}
	return name, nil
}
func (s *Store) get(id string) (gyrus.Concept, error) {
	name, err := s.getConceptPathByID(id)
	if err != nil {
		return gyrus.Concept{}, err
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return gyrus.Concept{}, err
	}
	doc, err := okf.ParseMarkdown(data)
	if err != nil {
		return gyrus.Concept{}, err
	}
	if doc.ID != filepath.Base(id) {
		return gyrus.Concept{}, fmt.Errorf("concept ID does not match filename: %s", name)
	}
	if err := okf.SetLocation(doc, filepath.ToSlash(name)); err != nil {
		return gyrus.Concept{}, err
	}
	if err := okf.Validate(doc); err != nil {
		return gyrus.Concept{}, err
	}
	return *doc, nil
}

func ref(doc gyrus.Concept) gyrus.ConceptRef {
	return gyrus.ConceptRef{Name: doc.Name, ID: doc.ID, Version: doc.Version, Status: doc.Status, LastUpdated: doc.LastUpdated}
}

func (s *Store) Create(ctx context.Context, doc gyrus.Concept) (gyrus.ConceptRef, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := fileio.Lock(ctx, s.rootDir)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}
	defer unlock()
	if err := okf.SetName(&doc, s.defaultProject); err != nil {
		return gyrus.ConceptRef{}, err
	}
	if err := okf.Validate(&doc); err != nil {
		return gyrus.ConceptRef{}, err
	}
	if _, err := s.getConceptPathByID(doc.Name); err == nil {
		return gyrus.ConceptRef{}, &gyrus.ConflictError{ConceptID: doc.Name, Operation: "create", Err: os.ErrExist}
	} else if !errors.Is(err, os.ErrNotExist) {
		return gyrus.ConceptRef{}, err
	}
	relative, err := okf.ConceptPath(doc, s.defaultProject)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}
	doc.LastUpdated = time.Now().UTC().Truncate(time.Second)
	data, err := okf.SerializeMarkdown(&doc)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}
	if err := fileio.Write(filepath.Join(s.rootDir, filepath.FromSlash(relative)), data); err != nil {
		return gyrus.ConceptRef{}, err
	}
	return ref(doc), nil
}

func (s *Store) Get(ctx context.Context, id string) (gyrus.Concept, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := fileio.Lock(ctx, s.rootDir)
	if err != nil {
		return gyrus.Concept{}, err
	}
	defer unlock()
	return s.get(id)
}

func (s *Store) Update(ctx context.Context, id string, patch gyrus.ConceptPatch, version int) (gyrus.ConceptRef, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := fileio.Lock(ctx, s.rootDir)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}
	defer unlock()
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
	data, err := okf.SerializeMarkdown(&doc)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}
	if err := fileio.Write(filepath.FromSlash(doc.FilePath), data); err != nil {
		return gyrus.ConceptRef{}, err
	}
	return ref(doc), nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := fileio.Lock(ctx, s.rootDir)
	if err != nil {
		return err
	}
	defer unlock()
	name, err := s.getConceptPathByID(id)
	if err != nil {
		return err
	}
	return os.Remove(name)
}
func (s *Store) Archive(ctx context.Context, id string) error { return s.Delete(ctx, id) }

// Scan takes a detached snapshot under the same lock used by mutations.
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
	unlock, err := fileio.Lock(ctx, s.rootDir)
	if err != nil {
		return nil, report, err
	}
	defer unlock()
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
		data, err := os.ReadFile(name)
		if err != nil {
			return nil, report, err
		}
		doc, err := okf.ParseMarkdown(data)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		if err := okf.SetLocation(doc, filepath.ToSlash(name)); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		if doc.ID+".md" != filepath.Base(name) {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: concept ID does not match filename", name))
			continue
		}
		docs = append(docs, *doc)
	}
	return docs, report, nil
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
	release, err := fileio.Lock(ctx, s.rootDir)
	if err != nil {
		return err
	}
	defer release()

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

	if err := fileio.Write(targetPath, []byte(content)); err != nil {
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
	release, err := fileio.Lock(ctx, s.rootDir)
	if err != nil {
		return err
	}
	defer release()

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
