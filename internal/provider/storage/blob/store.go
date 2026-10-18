package blob

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/internal/persistence/fileio"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
	"gocloud.dev/gcerrors"
)

// Store persists canonical OKF collections with conditional object mutations.
type Store struct {
	bucket                               *blob.Bucket
	prefix, scheme, bucketName, fileRoot string
	mu                                   sync.Mutex
}

// Option configures capabilities of an injected bucket.
type Option func(*Store)

// WithFileRoot enables interprocess locking for an injected fileblob bucket.
// The caller must pass the same root used to open that bucket.
func WithFileRoot(root string) Option { return func(s *Store) { s.fileRoot = root; s.scheme = "file" } }
func NewStore(ctx context.Context, bucketURL, prefix string) (*Store, error) {
	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, fmt.Errorf("invalid blob bucket URL")
	}
	if err := validatePrefix(prefix); err != nil {
		return nil, err
	}
	bucket, err := blob.OpenBucket(ctx, bucketURL)
	if err != nil {
		return nil, fmt.Errorf("open %s bucket failed", u.Scheme)
	}
	s := NewStoreWithBucket(bucket, prefix)
	s.scheme, s.bucketName = u.Scheme, u.Host
	if u.Scheme == "file" {
		s.fileRoot = filepath.FromSlash(u.Path)
		if u.Host != "" {
			s.fileRoot = filepath.Join(string(filepath.Separator)+string(filepath.Separator)+u.Host, s.fileRoot)
		}
		s.fileRoot, err = filepath.EvalSymlinks(s.fileRoot)
		if err != nil {
			_ = bucket.Close()
			return nil, err
		}
	}
	return s, nil
}
func NewStoreWithBucket(bucket *blob.Bucket, prefix string, options ...Option) *Store {
	s := &Store{bucket: bucket, prefix: strings.Trim(prefix, "/")}
	for _, option := range options {
		option(s)
	}
	return s
}
func (s *Store) Close() error { return s.bucket.Close() }
func validatePrefix(prefix string) error {
	for _, segment := range strings.Split(strings.Trim(prefix, "/"), "/") {
		if segment != "" {
			if err := okf.ValidateSegment("blob.prefix", segment); err != nil {
				return err
			}
		}
	}
	return nil
}
func (s *Store) lock(ctx context.Context) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validatePrefix(s.prefix); err != nil {
		return nil, err
	}
	if s.fileRoot != "" {
		return fileio.Lock(ctx, s.fileRoot)
	}
	return func() {}, nil
}
func (s *Store) makeKey(doc gyrus.Concept) string {
	relative, _ := okf.ConceptPath(doc, "")
	return path.Join(s.prefix, relative)
}
func serializeConcept(doc *gyrus.Concept) ([]byte, error) { return okf.SerializeMarkdown(doc) }
func deserializeConcept(data []byte) (gyrus.Concept, error) {
	doc, err := okf.ParseMarkdown(data)
	if err != nil {
		return gyrus.Concept{}, err
	}
	return *doc, nil
}
func (s *Store) keys(ctx context.Context) ([]string, error) {
	prefix := path.Join(s.prefix, "concepts") + "/"
	iter := s.bucket.List(&blob.ListOptions{Prefix: prefix})
	var keys []string
	for {
		obj, err := iter.Next(ctx)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if !obj.IsDir && strings.HasSuffix(obj.Key, ".md") && !strings.HasPrefix(path.Base(obj.Key), ".") {
			keys = append(keys, obj.Key)
		}
	}
	sort.Strings(keys)
	return keys, nil
}
func (s *Store) findKeyByID(ctx context.Context, id string) (string, error) {
	if err := okf.ValidateName(id); err != nil {
		return "", err
	}
	key := path.Join(s.prefix, "concepts", id+".md")
	if _, err := s.bucket.Attributes(ctx, key); err != nil {
		return "", err
	}
	return key, nil
}
func (s *Store) read(ctx context.Context, key string) (gyrus.Concept, objectToken, error) {
	data, token, err := s.readBytes(ctx, key)
	if err != nil {
		return gyrus.Concept{}, token, err
	}
	doc, err := deserializeConcept(data)
	if err != nil {
		return doc, token, &okf.ValidationError{Field: "payload", Message: err.Error()}
	}
	if err := okf.SetLocation(&doc, key); err != nil {
		return doc, token, err
	}
	if path.Base(key) != doc.ID+".md" {
		return doc, token, &okf.ValidationError{Field: "id", Message: "concept ID does not match key: " + key}
	}
	if err := okf.Validate(&doc); err != nil {
		return doc, token, err
	}
	return doc, token, nil
}

func (s *Store) readBytes(ctx context.Context, key string) ([]byte, objectToken, error) {
	if s.fileRoot != "" {
		if _, err := fileio.SafePath(s.fileRoot, key); err != nil {
			return nil, objectToken{}, err
		}
	}
	reader, err := s.bucket.NewReader(ctx, key, nil)
	if err != nil {
		return nil, objectToken{}, err
	}
	defer reader.Close()
	token := readerToken(reader)
	data, err := io.ReadAll(reader)
	return data, token, err
}
func (s *Store) conflict(id, op string, err error) error {
	if gcerrors.Code(err) == gcerrors.FailedPrecondition {
		return &gyrus.ConflictError{ConceptID: id, Operation: op, Err: errors.New("object changed since it was read")}
	}
	return cloudConflict(id, op, err)
}
func blobRef(doc gyrus.Concept) gyrus.ConceptRef {
	return gyrus.ConceptRef{Name: doc.Name, ID: doc.ID, Version: doc.Version, Status: doc.Status, LastUpdated: doc.LastUpdated}
}
func (s *Store) Create(ctx context.Context, doc gyrus.Concept) (gyrus.ConceptRef, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lock(ctx)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}
	defer unlock()
	if err := okf.SetName(&doc, ""); err != nil {
		return gyrus.ConceptRef{}, err
	}
	if err := okf.Validate(&doc); err != nil {
		return gyrus.ConceptRef{}, err
	}
	if _, err := okf.ConceptPath(doc, ""); err != nil {
		return gyrus.ConceptRef{}, err
	}
	if _, err := s.findKeyByID(ctx, doc.Name); err == nil {
		return gyrus.ConceptRef{}, &gyrus.ConflictError{ConceptID: doc.Name, Operation: "create", Err: os.ErrExist}
	} else if !errors.Is(err, os.ErrNotExist) {
		return gyrus.ConceptRef{}, err
	}
	doc.LastUpdated = time.Now().UTC().Truncate(time.Second)
	data, err := serializeConcept(&doc)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}
	if err := s.writeConditional(ctx, s.makeKey(doc), data, objectToken{}, true); err != nil {
		return gyrus.ConceptRef{}, s.conflict(doc.Name, "create", err)
	}
	return blobRef(doc), nil
}
func (s *Store) Get(ctx context.Context, id string) (gyrus.Concept, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lock(ctx)
	if err != nil {
		return gyrus.Concept{}, err
	}
	defer unlock()
	key, err := s.findKeyByID(ctx, id)
	if err != nil {
		return gyrus.Concept{}, err
	}
	doc, _, err := s.read(ctx, key)
	return doc, err
}
func (s *Store) Update(ctx context.Context, id string, patch gyrus.ConceptPatch, version int) (gyrus.ConceptRef, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lock(ctx)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}
	defer unlock()
	key, err := s.findKeyByID(ctx, id)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}
	doc, token, err := s.read(ctx, key)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}
	if version > 0 && version != doc.Version {
		return gyrus.ConceptRef{}, &gyrus.ConflictError{ConceptID: id, Operation: "update", Err: fmt.Errorf("expected version %d, got %d", version, doc.Version)}
	}
	if err := lifecycle.ApplyPatch(&doc, patch); err != nil {
		return gyrus.ConceptRef{}, err
	}
	data, err := serializeConcept(&doc)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}
	if err := s.writeConditional(ctx, key, data, token, false); err != nil {
		return gyrus.ConceptRef{}, s.conflict(id, "update", err)
	}
	return blobRef(doc), nil
}
func (s *Store) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()
	key, err := s.findKeyByID(ctx, id)
	if err != nil {
		return err
	}
	_, token, err := s.read(ctx, key)
	if err != nil {
		return err
	}
	return s.conflict(id, "delete", s.deleteConditional(ctx, key, token))
}
func (s *Store) Archive(ctx context.Context, id string) error {
	doc, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	status := "archived"
	_, err = s.Update(ctx, id, gyrus.ConceptPatch{Status: &status}, doc.Version)
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
	unlock, err := s.lock(ctx)
	if err != nil {
		return nil, report, err
	}
	defer unlock()
	keys, err := s.keys(ctx)
	if err != nil {
		return nil, report, err
	}
	var docs []gyrus.Concept
	for _, key := range keys {
		report.ScannedFiles++
		doc, _, err := s.read(ctx, key)
		if err != nil {
			var validation *okf.ValidationError
			if errors.As(err, &validation) {
				report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", key, err))
				continue
			}
			return nil, report, err
		}
		docs = append(docs, doc)
	}
	return docs, report, nil
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
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()
	if err := okf.ValidateSchemaContent(conceptType, content); err != nil {
		return err
	}

	key := s.schemaKey(conceptType)
	if err := okf.ValidateSchemaPath(key); err != nil {
		return err
	}

	_, token, err := s.readBytes(ctx, key)
	create := gcerrors.Code(err) == gcerrors.NotFound || errors.Is(err, os.ErrNotExist)
	if err != nil && !create {
		return err
	}
	return s.conflict(conceptType, "save schema", s.writeConditional(ctx, key, []byte(content), token, create))
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
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()
	key := s.schemaKey(conceptType)
	if err := okf.ValidateSchemaPath(key); err != nil {
		return err
	}

	_, token, err := s.readBytes(ctx, key)
	if err != nil {
		return err
	}
	return s.conflict(conceptType, "delete schema", s.deleteConditional(ctx, key, token))
}
