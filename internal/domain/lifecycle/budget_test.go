package lifecycle_test

import (
	"context"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that SuggestContextWithBudget applies the token budget and that a zero budget keeps whole concepts.
// [Execution Surface]: Lifecycle engine over an in-memory store
// [Assertions]: With a budget the layer fits it and holds the matching section, not the filler; without one it holds
// the whole concept as SuggestContextWithFilter always has.
// -----------------------------------------------------------------------------
func TestEngineSuggestContextWithBudget(t *testing.T) {
	ctx := context.Background()
	filler := strings.Repeat("Filler sentence with no useful words. ", 60)
	doc := gyrus.Concept{
		ID: "runbook", Title: "Runbook", Category: gyrus.CategoryOperations, Type: gyrus.TypeGuide,
		Tenant: "root", Scope: "reference", Name: "root/reference/runbook",
		Content: "# Runbook\n\n## Setup\n\n" + filler + "\n\n## Restart the queue\n\nRun the queue restart script.\n\n## Teardown\n\n" + filler,
	}
	store := &mockStoreWithSearch{docs: map[string]gyrus.Concept{doc.ID: doc}}
	engine := lifecycle.NewEngine(store, store, nil, nil, "")

	layer, err := engine.SuggestContextWithBudget(ctx, "restart the queue", gyrus.SearchFilter{}, 5, 200)
	if err != nil {
		t.Fatalf("SuggestContextWithBudget failed: %v", err)
	}
	if lifecycle.EstimateTokensFromBytes(layer) > 200 {
		t.Errorf("layer of %d estimated tokens exceeds 200:\n%s", lifecycle.EstimateTokensFromBytes(layer), layer)
	}
	if !strings.Contains(layer, "## Restart the queue") || strings.Contains(layer, "Filler sentence") {
		t.Errorf("expected only the matching section:\n%s", layer)
	}

	whole, err := engine.SuggestContextWithFilter(ctx, "restart the queue", gyrus.SearchFilter{}, 5)
	if err != nil {
		t.Fatalf("SuggestContextWithFilter failed: %v", err)
	}
	if !strings.Contains(whole, doc.Content) {
		t.Errorf("a zero budget must return the whole concept:\n%s", whole)
	}
}
