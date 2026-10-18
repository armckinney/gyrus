package git

import (
	"context"
	"errors"
	"testing"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

func TestRejectedPushCannotLeakUnconfirmedState(t *testing.T) {
	ctx := context.Background()
	remote := setupTestRemote(t)
	first, err := NewStore(Options{RepoURL: remote})
	if err != nil {
		t.Fatal(err)
	}
	doc := gyrus.Concept{ID: "shared", Title: "original", Category: gyrus.CategoryTechnical, Type: gyrus.TypeGuide, Tenant: "root", Status: "active", Version: 1, Name: "root/reference/shared"}
	if _, err := first.Create(ctx, doc); err != nil {
		t.Fatal(err)
	}
	second, err := NewStore(Options{RepoURL: remote})
	if err != nil {
		t.Fatal(err)
	}
	stale, err := second.get("root/reference/shared")
	if err != nil {
		t.Fatal(err)
	}
	winner := "winner"
	if _, err := first.Update(ctx, "root/reference/shared", gyrus.ConceptPatch{Title: &winner}, 1); err != nil {
		t.Fatal(err)
	}
	stale.Title = "unconfirmed loser"
	stale.Version++
	if err := second.write(stale.FilePath, stale); err != nil {
		t.Fatal(err)
	}
	err = second.commit(ctx, "root/reference/shared", "losing commit", stale.FilePath, false)
	var conflict *gyrus.ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("push conflict: %T %v", err, err)
	}
	got, err := second.Get(ctx, "root/reference/shared")
	if err != nil || got.Title != "winner" || got.Version != 2 {
		t.Fatalf("unconfirmed data leaked: %+v %v", got, err)
	}
	if _, err := second.Update(ctx, "root/reference/shared", gyrus.ConceptPatch{Title: &winner}, 1); !errors.As(err, &conflict) {
		t.Fatalf("stale version accepted: %v", err)
	}
	count := 0
	report, err := second.Scan(ctx, func(d gyrus.Concept) error {
		count++
		if d.Scope != "reference" {
			t.Fatal(d.Scope)
		}
		return nil
	})
	if err != nil || count != 1 || len(report.Errors) > 0 {
		t.Fatalf("scan: %+v %v", report, err)
	}
}
