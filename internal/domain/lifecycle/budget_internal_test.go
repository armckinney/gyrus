package lifecycle

import (
	"fmt"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that concept content splits into sections at Markdown headings, never inside fenced code.
// [Execution Surface]: Pure function
// [Assertions]: Preamble, each heading level and CRLF content split as expected; '#' lines inside fences do not split.
// -----------------------------------------------------------------------------
func TestSplitSectionsAtHeadings(t *testing.T) {
	content := "Intro line.\r\n\r\n# Title\r\nText.\r\n\r\n## Sub\r\nMore.\r\n```sh\r\n# not a heading\r\n```\r\n#hashtag is text\r\n### Deep\r\nEnd."
	got := splitSections(content)
	want := []string{"", "Title", "Sub", "Deep"}
	if len(got) != len(want) {
		t.Fatalf("expected %d sections, got %d: %+v", len(want), len(got), got)
	}
	for i, s := range got {
		if s.heading != want[i] || s.order != i {
			t.Errorf("section %d: heading %q order %d, want %q", i, s.heading, s.order, want[i])
		}
	}
	if !strings.Contains(got[2].text, "# not a heading") || !strings.Contains(got[2].text, "#hashtag is text") {
		t.Errorf("fenced and non-heading '#' lines must stay in their section: %q", got[2].text)
	}
}

func budgetDocs() (project, ref []gyrus.Concept) {
	var body strings.Builder
	body.WriteString("# Deploy guide\n\nOverview of every deploy.\n\n")
	for i := 0; i < 12; i++ {
		fmt.Fprintf(&body, "## Step %d\n\n%s\n\n", i, strings.Repeat("Unrelated filler words about nothing in particular. ", 20))
	}
	body.WriteString("## Rollback\n\nTo roll back a release, redeploy the previous tag with the rollback script.\n")
	project = []gyrus.Concept{{Name: "root/projects/app/deploy", Title: "Deploy", Category: "operations", Type: "guide", Scope: "project", Project: "app", Content: body.String()}}
	ref = []gyrus.Concept{
		{Name: "root/reference/glossary", Title: "Glossary", Category: "technical", Type: "glossary", Scope: "reference", Content: "# Glossary\n\n## Release\n\nA tagged build.\n\n## Tag\n\nA git ref.\n"},
		{Name: "root/reference/big", Title: "Big", Category: "technical", Type: "guide", Scope: "reference", Content: "# Big\n\n" + strings.Repeat("A release paragraph that is long. ", 400)},
	}
	return project, ref
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that the budgeted context layer never exceeds its token budget and returns matching sections first.
// [Execution Surface]: Pure function
// [Assertions]: For budgets from 40 to 6000 the estimate stays within the budget; at a small budget the rollback
// section is chosen over filler and the omitted sections are noted; an oversized section is cut at a paragraph.
// -----------------------------------------------------------------------------
func TestRenderBudgetedStaysWithinBudget(t *testing.T) {
	project, ref := budgetDocs()
	for _, budget := range []int{40, 100, 250, 500, 1000, 2000, 6000} {
		out := renderBudgeted("how do I roll back a release", "[PROJECT CONTEXT: app]\n\n", "[SHARED REFERENCE LAYER: root]\n\n", project, ref, budget)
		if got := EstimateTokensFromBytes(out); got > budget {
			t.Errorf("budget %d: estimate %d exceeds it:\n%s", budget, got, out)
		}
	}

	out := renderBudgeted("how do I roll back a release", "[PROJECT CONTEXT: app]\n\n", "[SHARED REFERENCE LAYER: root]\n\n", project, ref, 300)
	for _, want := range []string{"## Rollback", "redeploy the previous tag", "## Release", "more sections: gyrus get root/projects/app/deploy"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in layer:\n%s", want, out)
		}
	}
	if strings.Contains(out, "## Step 3") {
		t.Errorf("filler sections must not displace matching ones:\n%s", out)
	}
	if strings.Index(out, "[PROJECT CONTEXT: app]") > strings.Index(out, "[SHARED REFERENCE LAYER: root]") {
		t.Errorf("project context must come first:\n%s", out)
	}

	big := renderBudgeted("release", "", "[SHARED REFERENCE LAYER: root]\n\n", nil, ref[1:], 500)
	if !strings.Contains(big, truncationMark) || EstimateTokensFromBytes(big) > 500 {
		t.Errorf("oversized section must be cut at a paragraph within budget (estimate %d):\n%s", EstimateTokensFromBytes(big), big)
	}
}
