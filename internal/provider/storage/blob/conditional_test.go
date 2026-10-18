package blob

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	gocloudblob "gocloud.dev/blob"
	"gocloud.dev/blob/azureblob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/memblob"
	"gocloud.dev/blob/s3blob"
	"google.golang.org/api/option"
)

func TestCloudConditionalMutations(t *testing.T) {
	for _, scheme := range []string{"s3", "azblob", "gs"} {
		t.Run(scheme, func(t *testing.T) {
			ctx := context.Background()
			const key = "concepts/root/reference/shared.md"
			payload := []byte("---\nid: shared\ntitle: Shared\ntenant: root\ncategory: technical\ntype: guide\nstatus: active\nversion: 1\n---\n\nwinner content\n")
			var mu sync.Mutex
			updateCondition, deleteCondition, createCondition := false, false, false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				if r.Method == http.MethodGet {
					w.Header().Set("ETag", `"observed"`)
					w.Header().Set("X-Goog-Generation", "7")
					w.Header().Set("X-Ms-Blob-Type", "BlockBlob")
					w.Header().Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 GMT")
					w.Header().Set("Content-Type", "text/markdown")
					w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
					_, _ = w.Write(payload)
					return
				}
				if scheme == "azblob" && r.URL.Query().Get("comp") == "block" {
					w.WriteHeader(http.StatusCreated)
					return
				}
				condition := r.Header.Get("If-Match")
				generation := r.URL.Query().Get("ifGenerationMatch")
				if r.Method == http.MethodDelete {
					deleteCondition = condition == `"observed"` || generation == "7"
				} else if condition != "" || generation == "7" {
					updateCondition = condition == `"observed"` || generation == "7"
				} else {
					createCondition = r.Header.Get("If-None-Match") == "*" || generation == "0"
				}
				w.Header().Set("Content-Type", "application/xml")
				if scheme == "gs" {
					w.Header().Set("Content-Type", "application/json")
				}
				w.WriteHeader(http.StatusPreconditionFailed)
				switch scheme {
				case "s3":
					fmt.Fprint(w, `<Error><Code>PreconditionFailed</Code><Message>competing writer won</Message></Error>`)
				case "azblob":
					w.Header().Set("x-ms-error-code", "ConditionNotMet")
					fmt.Fprint(w, `<Error><Code>ConditionNotMet</Code><Message>competing writer won</Message></Error>`)
				default:
					fmt.Fprint(w, `{"error":{"code":412,"message":"competing writer won"}}`)
				}
			}))
			defer server.Close()
			var bucket *gocloudblob.Bucket
			var err error
			switch scheme {
			case "s3":
				client := s3.NewFromConfig(aws.Config{Region: "us-east-1", Credentials: aws.AnonymousCredentials{}, HTTPClient: server.Client()}, func(o *s3.Options) { o.BaseEndpoint = &server.URL; o.UsePathStyle = true; o.RetryMaxAttempts = 1 })
				bucket, err = s3blob.OpenBucket(ctx, client, "bucket", nil)
			case "azblob":
				client, e := container.NewClientWithNoCredential(server.URL+"/bucket", &container.ClientOptions{ClientOptions: azcore.ClientOptions{Transport: server.Client(), Retry: policy.RetryOptions{MaxRetries: -1}}})
				if e != nil {
					t.Fatal(e)
				}
				bucket, err = azureblob.OpenBucket(ctx, client, nil)
			case "gs":
				client, e := storage.NewClient(ctx, option.WithEndpoint(server.URL+"/storage/v1/"), option.WithoutAuthentication(), option.WithHTTPClient(server.Client()), storage.WithJSONReads())
				if e != nil {
					t.Fatal(e)
				}
				defer client.Close()
				bucket, err = gcsblob.OpenBucket(ctx, nil, "bucket", &gcsblob.Options{Client: client})
			}
			if err != nil {
				t.Fatal(err)
			}
			defer bucket.Close()
			store := NewStoreWithBucket(bucket, "")
			store.scheme, store.bucketName = scheme, "bucket"
			doc, token, err := store.read(ctx, key)
			if err != nil {
				t.Fatal(err)
			}
			if doc.Content != "winner content" {
				t.Fatal(doc.Content)
			}
			if scheme == "gs" && token.generation != 7 || scheme != "gs" && token.etag != `"observed"` {
				t.Fatalf("reader-bound token: %+v", token)
			}
			for _, op := range []string{"update", "create", "delete"} {
				switch op {
				case "update":
					err = store.conflict("root/reference/shared", op, store.writeConditional(ctx, key, []byte("loser"), token, false))
				case "create":
					err = store.conflict("root/reference/shared", op, store.writeConditional(ctx, key, []byte("loser"), objectToken{}, true))
				case "delete":
					err = store.deleteConditional(ctx, key, token)
				}
				var conflict *gyrus.ConflictError
				if !errors.As(err, &conflict) {
					t.Fatalf("%s did not return conflict: %T %v", op, err, err)
				}
			}
			mu.Lock()
			defer mu.Unlock()
			if !updateCondition || !createCondition || !deleteCondition {
				t.Fatalf("conditions: update=%v create=%v delete=%v", updateCondition, createCondition, deleteCondition)
			}
		})
	}
}

func TestOpaqueBucketRejectsUnsafeMutation(t *testing.T) {
	bucket := memblob.OpenBucket(nil)
	defer bucket.Close()
	store := NewStoreWithBucket(bucket, "")
	err := store.writeConditional(context.Background(), "a.md", []byte("bad"), objectToken{}, true)
	var unsupported *gyrus.UnsupportedConditionalMutationError
	if !errors.As(err, &unsupported) {
		t.Fatalf("unsafe mutation accepted: %v", err)
	}
	if exists, _ := bucket.Exists(context.Background(), "a.md"); exists {
		t.Fatal("opaque mutation wrote bytes")
	}
	if _, err := NewStore(context.Background(), "file:///unused", "../escape"); err == nil || !strings.Contains(err.Error(), "blob.prefix") {
		t.Fatalf("prefix validation: %v", err)
	}
}
