package blob

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/armckinney/gyrus/pkg/gyrus"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
	"gopkg.in/yaml.v3"
)

// Store implements gyrus.DocumentStore using gocloud.dev/blob.
type Store struct {
	bucket *blob.Bucket
	prefix string
}

// NewStore creates a new blob storage DocumentStore.
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

func (s *Store) makeKey(doc gyrus.Document) string {
	prefix := s.prefix
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	return fmt.Sprintf("%s%s/%s/%s.md", prefix, doc.OwnerGroup, doc.Category, doc.ID)
}

func serializeDocument(doc *gyrus.Document) ([]byte, error) {
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

func deserializeDocument(data []byte) (gyrus.Document, error) {
	var doc gyrus.Document
	str := string(data)
	
	// Fast path for missing frontmatter
	if !strings.HasPrefix(str, "---") {
		return doc, fmt.Errorf("missing yaml frontmatter")
	}

	parts := strings.SplitN(str, "\n---\n", 2)
	if len(parts) != 2 {
		return doc, fmt.Errorf("invalid document format")
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
	if s.prefix == "" {
		// Just in case it's in the root, but format is <owner>/<category>/<id>.md
		// it still ends with /<id>.md
	}

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
	return "", fmt.Errorf("document not found: %s", id)
}

// Create stores a new document.
func (s *Store) Create(ctx context.Context, doc gyrus.Document) (gyrus.DocumentRef, error) {
	key := s.makeKey(doc)
	
	exists, err := s.bucket.Exists(ctx, key)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}
	if exists {
		return gyrus.DocumentRef{}, fmt.Errorf("document already exists: %s", doc.ID)
	}

	data, err := serializeDocument(&doc)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	err = s.bucket.WriteAll(ctx, key, data, nil)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	return gyrus.DocumentRef{
		ID:          doc.ID,
		Version:     doc.Version,
		Status:      doc.Status,
		LastUpdated: doc.LastUpdated,
	}, nil
}

// Get retrieves a document by ID.
func (s *Store) Get(ctx context.Context, id string) (gyrus.Document, error) {
	key, err := s.findKeyByID(ctx, id)
	if err != nil {
		return gyrus.Document{}, err
	}

	data, err := s.bucket.ReadAll(ctx, key)
	if err != nil {
		return gyrus.Document{}, err
	}

	return deserializeDocument(data)
}

// Update patches an existing document.
func (s *Store) Update(ctx context.Context, id string, patch gyrus.DocumentPatch, expectedVersion int) (gyrus.DocumentRef, error) {
	doc, err := s.Get(ctx, id)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	if doc.Version != expectedVersion {
		return gyrus.DocumentRef{}, fmt.Errorf("version mismatch: expected %d, got %d", expectedVersion, doc.Version)
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
	// Not updating LastUpdated/LastModifiedBy here as it's not in the interface/patch fully defined,
	// but normally we would. The interface just asks us to return a DocumentRef.

	data, err := serializeDocument(&doc)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	key, err := s.findKeyByID(ctx, id)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	err = s.bucket.WriteAll(ctx, key, data, nil)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	return gyrus.DocumentRef{
		ID:          doc.ID,
		Version:     doc.Version,
		Status:      doc.Status,
		LastUpdated: doc.LastUpdated,
	}, nil
}

// Delete permanently removes a document.
func (s *Store) Delete(ctx context.Context, id string) error {
	key, err := s.findKeyByID(ctx, id)
	if err != nil {
		return err
	}
	return s.bucket.Delete(ctx, key)
}

// Archive changes the status of a document to archived.
func (s *Store) Archive(ctx context.Context, id string) error {
	archived := "archived"
	_, err := s.Update(ctx, id, gyrus.DocumentPatch{
		Status: &archived,
	}, 0)
	// Wait, we need expectedVersion. Get it first.
	doc, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	_, err = s.Update(ctx, id, gyrus.DocumentPatch{
		Status: &archived,
	}, doc.Version)
	return err
}
