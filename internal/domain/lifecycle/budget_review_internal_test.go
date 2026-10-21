package lifecycle

import (
	"fmt"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies the second review's budget fixes.
// [Execution Surface]: Pure function
// [Assertions]: a cut never splits a code block; "```a``` b" is not a fence; a concept with no content keeps its
// banner; a prompt of stopwords alone still scores sections; at a small budget the top match is opened before
// lower-ranked concepts with shorter banners.
// -----------------------------------------------------------------------------
func TestRenderBudgetedReviewFixes(t *testing.T) {
	code := "## Setup\n\nRun this:\n\n```sh\nstep one\n\nstep two\n```\n\n" + strings.Repeat("Filler words about setup. ", 60)
	for _, budget := range []int{60, 80, 100, 140} {
		out := renderBudgeted("setup", "", refBannerT, nil, []gyrus.Concept{refDoc("root/reference/code", code)}, budget)
		if strings.Count(out, "```")%2 != 0 {
			t.Errorf("budget %d: a cut left a code block open:\n%s", budget, out)
		}
	}
	if paras := paragraphs("a\n\n```\nx\n\ny\n```\n\nb"); len(paras) != 3 || !strings.Contains(paras[1], "x\n\ny") {
		t.Errorf("paragraphs must not split a code block: %q", paras)
	}
	if paras := paragraphs("## H\n\nbody\n\nmore"); len(paras) != 2 || paras[0] != "## H\n\nbody" {
		t.Errorf("a heading must stay with the paragraph after it: %q", paras)
	}
	giant := renderBudgeted("release", "", refBannerT, nil, []gyrus.Concept{refDoc("root/reference/giant", "# Giant\n\n"+strings.Repeat("One endless release paragraph. ", 300))}, 200)
	if !strings.Contains(giant, "more concepts matched: gyrus get root/reference/giant") {
		t.Errorf("a concept whose first paragraph cannot fit must be named, not cut mid-paragraph:\n%s", giant)
	}

	inline := splitSections("# A\n```npm install``` is the command\n## B\ntext")
	if len(inline) != 2 || inline[1].heading != "B" {
		t.Errorf("an inline code line must not open a fence: %+v", inline)
	}

	empty := renderBudgeted("queue", "", refBannerT, nil, []gyrus.Concept{refDoc("root/reference/full", "# Queue\n\nRestart the queue."), refDoc("root/reference/empty", "")}, 300)
	if !strings.Contains(empty, "--- [root/reference/empty]") {
		t.Errorf("a concept with no content must keep its banner:\n%s", empty)
	}

	stop := renderBudgeted("what is this", "", refBannerT, nil, []gyrus.Concept{refDoc("root/reference/a", "# Intro\n\nNothing here.\n\n# What this is\n\nThis is the answer.")}, 300)
	if !strings.Contains(stop, "This is the answer.") {
		t.Errorf("a prompt of stopwords alone must still pick the matching section:\n%s", stop)
	}

	// The live case: the top match has the longest banner, and five even shares are each too small for it.
	long := refDoc("root/reference/shared--decision-ledger-postgres-container", "# The ledger is its own Postgres container, with backups on N:\n\n## Decision\n\nThe ledger backups are kept on N:, dumped nightly.\n\n## Why\n\n"+strings.Repeat("Reasons. ", 40))
	long.Title = "The ledger is its own Postgres container, with backups on N:"
	ranked := []gyrus.Concept{long}
	for i := 0; i < 4; i++ {
		ranked = append(ranked, refDoc(fmt.Sprintf("root/reference/s%d", i), "# S\n\nledger "+strings.Repeat("word ", 30)))
	}
	for _, budget := range []int{150, 200, 300, 400} {
		out := renderBudgeted("where are the ledger backups kept", "", refBannerT, nil, ranked, budget)
		if EstimateTokensFromBytes(out) > budget {
			t.Fatalf("budget %d exceeded: %d", budget, EstimateTokensFromBytes(out))
		}
		if !strings.Contains(out, "--- ["+long.Name+"]") || !strings.Contains(out, "The ledger backups are kept on N:") {
			t.Errorf("budget %d: the top match must be opened with its answer:\n%s", budget, out)
		}
	}
}
