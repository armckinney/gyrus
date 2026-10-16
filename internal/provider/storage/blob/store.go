package blob

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
	"gopkg.in/yaml.v3"
)

// Store implements gyrus.ConceptStore using gocloud.dev/blob.
type Store struct {
	bucket *blob.Bucket
	prefix string
}

// NewStore creates a new blob storage ConceptStore.
func NewStore(ctx context.Context, bucketURL string, prefix string) (*Store, error) {
	bucket, err := blob.OpenBucket(ctx, bucketURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %q: %w", bucketURL, err)
	}
	return &Store{
		bucket: bucket,
		prefix: prefix,
	}, nil
}

// NewStoreWithBucket creates a store with an existing bucket.
func NewStoreWithBucket(bucket *blob.Bucket, prefix string) *Store {
	return &Store{
		bucket: bucket,
		prefix: prefix,
	}
}

// Close closes the underlying bucket.
func (s *Store) Close() error {
	return s.bucket.Close()
}

func (s *Store) makeKey(doc gyrus.Concept) string {
	prefix := s.prefix
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	return fmt.Sprintf("%s%s/%s/%s.md", prefix, doc.Tenant, doc.Category, doc.ID)
}

func serializeConcept(doc *gyrus.Concept) ([]byte, error) {
	data, err := yaml.Marshal(doc)
	if err != nil {
		return nil, err
	}
	content := []byte("---\n")
	content = append(content, data...)
	content = append(content, "---\n\n"...)
	content = append(content, []byte(doc.Content)...)
	return content, nil
}

func deserializeConcept(data []byte) (gyrus.Concept, error) {
	var doc gyrus.Concept
	str := string(data)

	// Fast path for missing frontmatter
	if !strings.HasPrefix(str, "---") {
		return doc, fmt.Errorf("missing yaml frontmatter")
	}

	parts := strings.SplitN(str, "\n---\n", 2)
	if len(parts) != 2 {
		return doc, fmt.Errorf("invalid concept format")
	}

	frontmatter := parts[0]
	frontmatter = strings.TrimPrefix(frontmatter, "---\n")

	if err := yaml.Unmarshal([]byte(frontmatter), &doc); err != nil {
		return doc, err
	}

	content := parts[1]
	doc.Content = strings.TrimPrefix(content, "\n")

	return doc, nil
}

func (s *Store) findKeyByID(ctx context.Context, id string) (string, error) {
	iter := s.bucket.List(&blob.ListOptions{Prefix: s.prefix})
	targetSuffix := "/" + id + ".md"

	for {
		obj, err := iter.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if strings.HasSuffix(obj.Key, targetSuffix) {
			return obj.Key, nil
		}
	}
	return "", fmt.Errorf("concept not found: %s", id)
}

// Create stores a new concept.
func (s *Store) Create(ctx context.Context, doc gyrus.Concept) (gyrus.ConceptRef, error) {
	key := s.makeKey(doc)

	exists, err := s.bucket.Exists(ctx, key)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}
	if exists {
		return gyrus.ConceptRef{}, fmt.Errorf("concept already exists: %s", doc.ID)
	}

	data, err := serializeConcept(&doc)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}

	err = s.bucket.WriteAll(ctx, key, data, nil)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}

	return gyrus.ConceptRef{
		ID:          doc.ID,
		Version:     doc.Version,
		Status:      doc.Status,
		LastUpdated: doc.LastUpdated,
	}, nil
}

// Get retrieves a concept by ID.
func (s *Store) Get(ctx context.Context, id string) (gyrus.Concept, error) {
	key, err := s.findKeyByID(ctx, id)
	if err != nil {
		return gyrus.Concept{}, err
	}

	data, err := s.bucket.ReadAll(ctx, key)
	if err != nil {
		return gyrus.Concept{}, err
	}

	return deserializeConcept(data)
}

// Update patches an existing concept.
func (s *Store) Update(ctx context.Context, id string, patch gyrus.ConceptPatch, expectedVersion int) (gyrus.ConceptRef, error) {
	doc, err := s.Get(ctx, id)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}

	if doc.Version != expectedVersion {
		return gyrus.ConceptRef{}, fmt.Errorf("version mismatch: expected %d, got %d", expectedVersion, doc.Version)
	}

	// Apply patch
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

	data, err := serializeConcept(&doc)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}

	key, err := s.findKeyByID(ctx, id)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}

	err = s.bucket.WriteAll(ctx, key, data, nil)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}

	return gyrus.ConceptRef{
		ID:          doc.ID,
		Version:     doc.Version,
		Status:      doc.Status,
		LastUpdated: doc.LastUpdated,
	}, nil
}

// Delete permanently removes a concept.
func (s *Store) Delete(ctx context.Context, id string) error {
	key, err := s.findKeyByID(ctx, id)
	if err != nil {
		return err
	}
	return s.bucket.Delete(ctx, key)
}

// Archive changes the status of a concept to archived.
func (s *Store) Archive(ctx context.Context, id string) error {
	archived := "archived"
	doc, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	_, err = s.Update(ctx, id, gyrus.ConceptPatch{
		Status: &archived,
	}, doc.Version)
	return err
}

func (s *Store) schemaPrefix() string {
	prefix := s.prefix
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	return fmt.Sprintf("%s.gyrus/schemas/", prefix)
}

func (s *Store) schemaKey(conceptType string) string {
	return fmt.Sprintf("%s%s.md", s.schemaPrefix(), conceptType)
}

// GetSchema retrieves a persisted schema template for a concept type from blob storage.
func (s *Store) GetSchema(ctx context.Context, conceptType string) (string, error) {
	key := s.schemaKey(conceptType)
	if err := okf.ValidateSchemaPath(key); err != nil {
		return "", err
	}

	data, err := s.bucket.ReadAll(ctx, key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveSchema persists an OKF schema template to the enforced .gyrus/schemas/ path in blob storage.
func (s *Store) SaveSchema(ctx context.Context, conceptType string, content string) error {
	if err := okf.ValidateSchemaContent(conceptType, content); err != nil {
		return err
	}

	key := s.schemaKey(conceptType)
	if err := okf.ValidateSchemaPath(key); err != nil {
		return err
	}

	return s.bucket.WriteAll(ctx, key, []byte(content), &blob.WriterOptions{
		ContentType: "text/markdown",
	})
}

// ListSchemas enumerates all persisted schema template types under .gyrus/schemas/ in blob storage.
func (s *Store) ListSchemas(ctx context.Context) ([]string, error) {
	prefix := s.schemaPrefix()
	iter := s.bucket.List(&blob.ListOptions{
		Prefix: prefix,
	})

	var list []string
	for {
		obj, err := iter.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		name := strings.TrimPrefix(obj.Key, prefix)
		if strings.HasSuffix(name, ".md") {
			list = append(list, strings.TrimSuffix(name, ".md"))
		}
	}
	sort.Strings(list)
	return list, nil
}

// DeleteSchema removes a persisted schema template from blob storage.
func (s *Store) DeleteSchema(ctx context.Context, conceptType string) error {
	key := s.schemaKey(conceptType)
	if err := okf.ValidateSchemaPath(key); err != nil {
		return err
	}

	return s.bucket.Delete(ctx, key)
}
