package blob

import (
	"context"
	"errors"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	azblobblob "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/armckinney/gyrus/internal/provider/storage/fileio"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gocloud.dev/blob"
	"google.golang.org/api/googleapi"
)

type objectToken struct {
	etag       string
	generation int64
}

func readerToken(reader *blob.Reader) objectToken {
	var aws s3.GetObjectOutput
	if reader.As(&aws) && aws.ETag != nil {
		return objectToken{etag: *aws.ETag}
	}
	var azure azblobblob.DownloadStreamResponse
	if reader.As(&azure) && azure.ETag != nil {
		return objectToken{etag: string(*azure.ETag)}
	}
	var google *storage.Reader
	if reader.As(&google) {
		return objectToken{generation: google.Attrs.Generation}
	}
	return objectToken{}
}
func (s *Store) unsupported(op string) error {
	driver := s.scheme
	if driver == "" {
		driver = "opaque blob bucket"
	}
	return &gyrus.UnsupportedConditionalMutationError{Driver: driver, Operation: op}
}
func (s *Store) writeConditional(ctx context.Context, key string, data []byte, token objectToken, create bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.fileRoot != "" {
		name, err := fileio.SafePath(s.fileRoot, key)
		if err != nil {
			return err
		}
		if create {
			if _, err := os.Stat(name); err == nil {
				return &gyrus.ConflictError{ConceptID: key, Operation: "create", Err: os.ErrExist}
			} else if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		return fileio.Write(name, data)
	}
	if s.scheme != "s3" && s.scheme != "azblob" && s.scheme != "gs" {
		return s.unsupported("write")
	}
	if !create && token.etag == "" && token.generation == 0 {
		return s.unsupported("update without object token")
	}
	opts := &blob.WriterOptions{IfNotExist: create, ContentType: "text/markdown"}
	if !create {
		opts.BeforeWrite = func(as func(any) bool) error {
			switch s.scheme {
			case "s3":
				var input *transfermanager.UploadObjectInput
				if as(&input) && token.etag != "" {
					input.IfMatch = &token.etag
					return nil
				}
			case "azblob":
				var input *azblob.UploadStreamOptions
				if as(&input) && token.etag != "" {
					etag := azcore.ETag(token.etag)
					input.AccessConditions = &azblobblob.AccessConditions{ModifiedAccessConditions: &azblobblob.ModifiedAccessConditions{IfMatch: &etag}}
					return nil
				}
			case "gs":
				var handle **storage.ObjectHandle
				if as(&handle) && token.generation > 0 {
					*handle = (*handle).If(storage.Conditions{GenerationMatch: token.generation})
					return nil
				}
			}
			return s.unsupported("conditional write hook")
		}
	}
	return s.bucket.WriteAll(ctx, key, data, opts)
}
func (s *Store) deleteConditional(ctx context.Context, key string, token objectToken) error {
	if s.fileRoot != "" {
		return s.bucket.Delete(ctx, key)
	}
	switch s.scheme {
	case "s3":
		var client *s3.Client
		if !s.bucket.As(&client) || token.etag == "" {
			return s.unsupported("delete")
		}
		_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &s.bucketName, Key: &key, IfMatch: &token.etag})
		return cloudConflict(key, "delete", err)
	case "azblob":
		var client *container.Client
		if !s.bucket.As(&client) || token.etag == "" {
			return s.unsupported("delete")
		}
		etag := azcore.ETag(token.etag)
		_, err := client.NewBlobClient(key).Delete(ctx, &azblobblob.DeleteOptions{AccessConditions: &azblobblob.AccessConditions{ModifiedAccessConditions: &azblobblob.ModifiedAccessConditions{IfMatch: &etag}}})
		return cloudConflict(key, "delete", err)
	case "gs":
		var client *storage.Client
		if !s.bucket.As(&client) || token.generation <= 0 {
			return s.unsupported("delete")
		}
		return cloudConflict(key, "delete", client.Bucket(s.bucketName).Object(key).If(storage.Conditions{GenerationMatch: token.generation}).Delete(ctx))
	}
	return s.unsupported("delete")
}

func cloudConflict(id, operation string, err error) error {
	if err == nil {
		return nil
	}
	var azure *azcore.ResponseError
	var google *googleapi.Error
	var httpError interface{ HTTPStatusCode() int }
	status := 0
	if errors.As(err, &azure) {
		status = azure.StatusCode
	} else if errors.As(err, &google) {
		status = google.Code
	} else if errors.As(err, &httpError) {
		status = httpError.HTTPStatusCode()
	}
	if status == http.StatusPreconditionFailed || status == http.StatusConflict {
		return &gyrus.ConflictError{ConceptID: id, Operation: operation, Err: errors.New("object changed since it was read")}
	}
	return err
}
