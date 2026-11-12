package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/armckinney/gyrus/pkg/gyrus"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func branchDoc(id string) gyrus.Concept {
	return gyrus.Concept{ID: id, Tenant: "root", Title: id, Category: gyrus.CategoryTechnical, Type: gyrus.TypeGuide, Status: "active"}
}

func TestInitiativesPersistAndIsolateAllMutations(t *testing.T) {
	ctx := context.Background()
	remote := setupTestRemote(t)
	trunk, err := NewStore(Options{RepoURL: remote})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = trunk.Create(ctx, branchDoc("shared")); err != nil {
		t.Fatal(err)
	}
	repo, _ := gogit.PlainOpen(remote)
	before, _ := repo.Reference(plumbing.NewBranchReferenceName("main"), true)
	session, err := NewStore(Options{RepoURL: remote, CommitStrategy: "branching"})
	if err != nil {
		t.Fatal(err)
	}
	if !session.Status().SelectionRequired {
		t.Fatal("selection should be required")
	}
	if _, err = session.Get(ctx, "root/reference/shared"); err == nil {
		t.Fatal("unselected session read knowledge")
	}
	for _, branch := range []string{"initiative/a", "initiative/b"} {
		if _, err = session.Switch(ctx, "branching", branch); err != nil {
			t.Fatal(err)
		}
		if _, err = session.Get(ctx, "root/reference/shared"); err != nil {
			t.Fatal(err)
		}
		if _, err = repo.Reference(plumbing.NewBranchReferenceName(branch), true); !errors.Is(err, plumbing.ErrReferenceNotFound) {
			t.Fatalf("read created remote branch: %v", err)
		}
		title := branch
		if _, err = session.Update(ctx, "root/reference/shared", gyrus.ConceptPatch{Title: &title, Links: &[]gyrus.ConceptLink{{Concept: "root/reference/future", Rel: gyrus.RelDependsOn}}}, 1); err != nil {
			t.Fatal(err)
		}
		schema := "---\nid: <id>\ntitle: " + branch + "\ntype: guide\n---\n# " + branch + "\n"
		if err = session.SaveSchema(ctx, "guide", schema); err != nil {
			t.Fatal(err)
		}
		if err = session.SaveSchema(ctx, "guide", schema); err != nil {
			t.Fatalf("unchanged schema: %v", err)
		}
		if _, err = session.Create(ctx, branchDoc("temporary")); err != nil {
			t.Fatal(err)
		}
		if err = session.Archive(ctx, "root/reference/temporary"); err != nil {
			t.Fatal(err)
		}
		if err = session.Delete(ctx, "root/reference/temporary"); err != nil {
			t.Fatal(err)
		}
	}
	for _, branch := range []string{"initiative/a", "initiative/b"} {
		fresh, err := NewStore(Options{RepoURL: remote, CommitStrategy: "branching", Branch: branch})
		if err != nil {
			t.Fatal(err)
		}
		doc, err := fresh.Get(ctx, "root/reference/shared")
		if err != nil || doc.Title != branch || len(doc.Links) != 1 {
			t.Fatalf("reload: %+v %v", doc, err)
		}
		schema, err := fresh.GetSchema(ctx, "guide")
		if err != nil || !strings.Contains(schema, branch) {
			t.Fatalf("schema isolation: %v", err)
		}
		if err = fresh.DeleteSchema(ctx, "guide"); err != nil {
			t.Fatal(err)
		}
	}
	after, _ := repo.Reference(plumbing.NewBranchReferenceName("main"), true)
	if before.Hash() != after.Hash() {
		t.Fatal("initiative writes modified trunk")
	}
	doc, err := trunk.Get(ctx, "root/reference/shared")
	if err != nil || doc.Title != "shared" || len(doc.Links) != 0 {
		t.Fatalf("trunk changed: %+v %v", doc, err)
	}
}

func TestSwitchFailureAndConcurrentSelections(t *testing.T) {
	ctx := context.Background()
	remote := setupTestRemote(t)
	s, err := NewStore(Options{RepoURL: remote})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Create(ctx, branchDoc("shared")); err != nil {
		t.Fatal(err)
	}
	before := s.Status()
	for _, tc := range []struct{ strategy, branch string }{{"unknown", ""}, {"branching", ""}, {"branching", "main"}, {"branching", "bad..name"}, {"direct", "initiative/a"}} {
		if _, err = s.Switch(ctx, tc.strategy, tc.branch); err == nil {
			t.Fatalf("accepted %+v", tc)
		}
		if s.Status() != before {
			t.Fatal("failed switch changed selection")
		}
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = s.Switch(canceled, "branching", "initiative/canceled"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if s.Status() != before {
		t.Fatal("canceled switch changed selection")
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, err := s.Switch(ctx, "branching", fmt.Sprintf("initiative/%d", i)); err != nil {
				t.Error(err)
			}
			s.Status()
			if _, err := s.Get(ctx, "root/reference/shared"); err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
}

func TestBranchingRequiresExistingTrunk(t *testing.T) {
	_, err := NewStore(Options{RepoURL: setupTestRemote(t), CommitStrategy: "branching", Branch: "initiative/a"})
	if err == nil {
		t.Fatal("branching initialized empty remote")
	}
}

func TestCredentialHelperPreservesCredentialsAndDisablesPrompts(t *testing.T) {
	dir := t.TempDir()
	helper := filepath.Join(dir, "git")
	script := "#!/bin/sh\n[ \"$1 $2\" = 'credential fill' ] || exit 4\n[ \"$GIT_TERMINAL_PROMPT\" = '0' ] || exit 5\n[ \"$GCM_INTERACTIVE\" = 'never' ] || exit 6\n[ -z \"$GIT_ASKPASS\" ] || exit 7\nread line\n[ \"$line\" = 'url=https://bitbucket.org/team/wiki.git' ] || exit 8\nprintf 'username=x-bitbucket-api-token-auth\\npassword=secret-with=equals\\n\\n'\n"
	if err := os.WriteFile(helper, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("GIT_AUTH_TOKEN", "ignored-token")
	t.Setenv("GIT_ASKPASS", "unexpected-prompt")
	auth, err := discoverAuth(context.Background(), "https://bitbucket.org/team/wiki.git")
	if err != nil {
		t.Fatal(err)
	}
	basic, ok := auth.(*http.BasicAuth)
	if !ok || basic.Username != "x-bitbucket-api-token-auth" || basic.Password != "secret-with=equals" {
		t.Fatal("helper credentials were not preserved")
	}
	s := &Store{repoURL: "https://bitbucket.org/team/wiki.git", auth: basic}
	if strings.Contains(s.remoteError("push", fmt.Errorf("rejected secret-with=equals")).Error(), basic.Password) {
		t.Fatal("secret leaked")
	}
	if err := os.WriteFile(helper, []byte("#!/bin/sh\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if auth, err = discoverAuth(context.Background(), s.repoURL); err != nil || auth != nil {
		t.Fatal("public anonymous fallback failed")
	}
	if err := os.WriteFile(helper, []byte("#!/bin/sh\nwhile :; do :; done\n"), 0700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err = discoverAuth(ctx, s.repoURL); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("helper cancellation: %v", err)
	}
}

func TestSchemaRejectedPushAndCancellationDoNotLeak(t *testing.T) {
	for _, strategy := range []string{"direct", "branching"} {
		t.Run(strategy, func(t *testing.T) {
			ctx := context.Background()
			remote := setupTestRemote(t)
			trunk, err := NewStore(Options{RepoURL: remote})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = trunk.Create(ctx, branchDoc("seed")); err != nil {
				t.Fatal(err)
			}
			opts := Options{RepoURL: remote, CommitStrategy: strategy}
			if strategy == "branching" {
				opts.Branch = "initiative/schema"
			}
			first, err := NewStore(opts)
			if err != nil {
				t.Fatal(err)
			}
			original := "---\nid: <id>\ntitle: Original\ntype: guide\n---\n# Original\n"
			winner := strings.ReplaceAll(original, "Original", "Winner")
			if err = first.SaveSchema(ctx, "guide", original); err != nil {
				t.Fatal(err)
			}
			second, err := NewStore(opts)
			if err != nil {
				t.Fatal(err)
			}
			if err = first.SaveSchema(ctx, "guide", winner); err != nil {
				t.Fatal(err)
			}
			path := second.schemaPath("guide")
			f, err := second.fs.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = f.Write([]byte(strings.ReplaceAll(original, "Original", "Unconfirmed"))); err != nil {
				t.Fatal(err)
			}
			if err = f.Close(); err != nil {
				t.Fatal(err)
			}
			err = second.commit(ctx, "", "stale schema", path, false)
			var conflict *gyrus.ConflictError
			if !errors.As(err, &conflict) {
				t.Fatalf("schema rejected push: %v", err)
			}
			actual, err := second.GetSchema(ctx, "guide")
			if err != nil || actual != winner {
				t.Fatalf("unconfirmed schema leaked: %q %v", actual, err)
			}
			canceled, cancel := context.WithCancel(ctx)
			cancel()
			if err = second.SaveSchema(canceled, "guide", original); !errors.Is(err, context.Canceled) {
				t.Fatalf("schema save cancellation: %v", err)
			}
			if err = second.DeleteSchema(canceled, "guide"); !errors.Is(err, context.Canceled) {
				t.Fatalf("schema delete cancellation: %v", err)
			}
			actual, err = second.GetSchema(ctx, "guide")
			if err != nil || actual != winner {
				t.Fatal("canceled operation changed schema")
			}
		})
	}
}
