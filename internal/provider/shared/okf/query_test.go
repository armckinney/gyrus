package okf_test

import (
	"context"
	"testing"

	sharedokf "github.com/armckinney/gyrus/internal/provider/shared/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

type scanner []gyrus.Concept

func (s scanner) Scan(ctx context.Context, visit func(gyrus.Concept) error) (gyrus.SyncReport, error) {
	for _, doc := range s {
		if err := visit(doc); err != nil {
			return gyrus.SyncReport{}, err
		}
	}
	return gyrus.SyncReport{}, nil
}

func concept(name, scope, project, title, content string) gyrus.Concept {
	return gyrus.Concept{Name: name, ID: name, Title: title, Tenant: "root", Scope: scope, Project: project, Content: content}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that ranking is BM25 and that the scope boost multiplies a match instead of acting as a term.
// [Execution Surface]: OKF search over an in-memory scanner
// [Assertions]: A reference concept matching the rare query terms outranks a project concept matching one common
// word; a project concept matching nothing is not returned; between equal matches the project concept ranks first.
// -----------------------------------------------------------------------------
func TestSearchRanksByBM25WithScopeBoostOnTheMatch(t *testing.T) {
	ctx := context.Background()
	docs := scanner{
		concept("root/projects/app/notes", "project", "app", "Session notes", "Every session starts here. Session, session."),
		concept("root/projects/app/unrelated", "project", "app", "Unrelated", "Nothing to see."),
		concept("root/reference/rollback", "reference", "", "Rollback runbook", "Roll back a release by redeploying the previous tag."),
		concept("root/reference/session-a", "reference", "", "Sessions", "A session is one window."),
		concept("root/reference/other", "reference", "", "Other", "Deploy the server."),
	}
	filter := gyrus.SearchFilter{Project: "app"}

	results, err := sharedokf.Search(ctx, docs, gyrus.SearchQuery{Query: "rollback a release in this session", Filter: filter})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 || results[0].Concept.Name != "root/reference/rollback" {
		t.Fatalf("expected the rollback runbook first, got %+v", results)
	}
	for _, r := range results {
		if r.Concept.Name == "root/projects/app/unrelated" || r.Concept.Name == "root/reference/other" {
			t.Errorf("a concept with no query term must not match: %s", r.Concept.Name)
		}
	}

	results, err = sharedokf.Search(ctx, scanner{
		concept("root/reference/same", "reference", "", "Same", "queue restart"),
		concept("root/projects/app/same", "project", "app", "Same", "queue restart"),
	}, gyrus.SearchQuery{Query: "queue restart", Filter: filter})
	if err != nil || len(results) != 2 || results[0].Concept.Project != "app" || results[0].MatchReason != "Project Match" {
		t.Fatalf("equal matches must rank the project concept first: %+v %v", results, err)
	}
}
