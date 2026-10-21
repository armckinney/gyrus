package lifecycle

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

const (
	refBannerT     = "[SHARED REFERENCE LAYER: root]\n\n"
	projectBannerT = "[PROJECT CONTEXT: app]\n\n"
)

func refDoc(name, content string) gyrus.Concept {
	return gyrus.Concept{Name: name, Title: name, Category: "technical", Type: "guide", Scope: "reference", Content: content}
}

// randomContent builds Markdown with headings of every level, empty and repeated headings, fences of both kinds and
// odd lengths, CRLF lines, multibyte text and long paragraphs.
func randomContent(r *rand.Rand) string {
	words := []string{"queue", "restart", "release", "tag", "émote", "日本語", "wallet", "ledger", "the", "C#", "x"}
	var b strings.Builder
	for i, n := 0, r.Intn(12); i < n; i++ {
		switch r.Intn(6) {
		case 0:
			fmt.Fprintf(&b, "%s %s\n\n", strings.Repeat("#", 1+r.Intn(6)), words[r.Intn(len(words))])
		case 1:
			b.WriteString("##\n\n")
		case 2:
			fence := strings.Repeat([]string{"`", "~"}[r.Intn(2)], 3+r.Intn(2))
			fmt.Fprintf(&b, "%sjs\n# not a heading\n```\n%s\n\n", fence, fence)
		case 3:
			b.WriteString("line one\r\nline two\r\n\r\n")
		default:
			for j, m := 0, 1+r.Intn(80); j < m; j++ {
				b.WriteString(words[r.Intn(len(words))] + " ")
			}
			b.WriteString("\n\n")
		}
	}
	return b.String()
}

// -----------------------------------------------------------------------------
// [Test Level]: Property Test
// [Purpose]: Verifies the budget invariant and determinism over random concepts, prompts and budgets.
// [Execution Surface]: Pure function, 3000 seeded random cases
// [Assertions]: for budgets 1 to 400 the layer's estimate never exceeds the budget, and the same input gives the same
// output twice.
// -----------------------------------------------------------------------------
func TestRenderBudgetedPropertyWithinBudgetAndDeterministic(t *testing.T) {
	r := rand.New(rand.NewSource(20261020))
	prompts := []string{"restart the queue", "release tag", "émote wallet", "", "日本語"}
	for i := 0; i < 3000; i++ {
		var project, ref []gyrus.Concept
		for j, n := 0, r.Intn(6); j < n; j++ {
			doc := refDoc(fmt.Sprintf("root/reference/d%d", j), randomContent(r))
			if r.Intn(3) == 0 {
				doc.Scope, doc.Name = "project", fmt.Sprintf("root/projects/app/d%d", j)
				project = append(project, doc)
			} else {
				ref = append(ref, doc)
			}
		}
		prompt, budget := prompts[r.Intn(len(prompts))], 1+r.Intn(400)
		out := renderBudgeted(prompt, projectBannerT, refBannerT, project, ref, budget)
		if got := EstimateTokensFromBytes(out); got > budget {
			t.Fatalf("case %d: estimate %d exceeds budget %d:\n%s", i, got, budget, out)
		}
		if again := renderBudgeted(prompt, projectBannerT, refBannerT, project, ref, budget); again != out {
			t.Fatalf("case %d: output is not deterministic", i)
		}
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies the edges of the budget: below the header, no concepts, no headings, multibyte text.
// [Execution Surface]: Pure function
// [Assertions]: a budget below the header gives ""; no concepts gives the header alone; a concept without headings is
// one section; multibyte text is estimated by bytes and stays within budget.
// -----------------------------------------------------------------------------
func TestRenderBudgetedEdges(t *testing.T) {
	header := EstimateTokensFromBytes(contextLayerHeader)
	if out := renderBudgeted("q", "", refBannerT, nil, []gyrus.Concept{refDoc("root/reference/a", "q")}, header-1); out != "" {
		t.Errorf("below the header: got %q", out)
	}
	if out := renderBudgeted("q", "", refBannerT, nil, nil, 100); out != contextLayerHeader {
		t.Errorf("no concepts: got %q", out)
	}
	plain := renderBudgeted("queue", "", refBannerT, nil, []gyrus.Concept{refDoc("root/reference/a", "No heading at all, the queue restarts.")}, 200)
	if !strings.Contains(plain, "No heading at all, the queue restarts.") {
		t.Errorf("a concept without headings must be one section:\n%s", plain)
	}
	multi := strings.Repeat("日本語のテキスト ", 200)
	if out := renderBudgeted("日本語のテキスト", "", refBannerT, nil, []gyrus.Concept{refDoc("root/reference/m", "# 見出し\n\n"+multi)}, 120); EstimateTokensFromBytes(out) > 120 {
		t.Errorf("multibyte layer over budget: %d", EstimateTokensFromBytes(out))
	}
	if EstimateTokensFromBytes("日") != 2 || EstimateTokensFromBytes("") != 0 || EstimateTokensFromBytes("abc") != 2 {
		t.Errorf("EstimateTokensFromBytes must be ceil(bytes/2.6): got %d %d %d", EstimateTokensFromBytes("日"), EstimateTokensFromBytes(""), EstimateTokensFromBytes("abc"))
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that one concept's large best section cannot crowd out the concepts after it.
// [Execution Surface]: Pure function
// [Assertions]: with a first concept whose best section alone exceeds the budget, all five concepts are opened; the
// first is cut to an even share and marked; a concept that cannot be opened at all is named in the dropped note.
// -----------------------------------------------------------------------------
func TestRenderBudgetedDoesNotStarveLaterConcepts(t *testing.T) {
	big := "# Big\n\n" + strings.Repeat("queue restart paragraph with words. ", 30) + "\n\n" + strings.Repeat("more queue restart text here. ", 300)
	ref := []gyrus.Concept{refDoc("root/reference/big", big)}
	for i := 1; i < 5; i++ {
		ref = append(ref, refDoc(fmt.Sprintf("root/reference/small%d", i), fmt.Sprintf("# Small %d\n\nRestart the queue, step %d.", i, i)))
	}
	out := renderBudgeted("restart the queue", "", refBannerT, nil, ref, 600)
	for _, doc := range ref {
		if !strings.Contains(out, "--- ["+doc.Name+"]") {
			t.Errorf("%s was crowded out:\n%s", doc.Name, out)
		}
	}
	if !strings.Contains(out, truncationMark) {
		t.Errorf("the large section must be cut to its share:\n%s", out)
	}

	many := []gyrus.Concept{}
	for i := 0; i < 5; i++ {
		many = append(many, refDoc(fmt.Sprintf("root/reference/n%d", i), "# N\n\n"+strings.Repeat("queue words ", 40)))
	}
	tight := renderBudgeted("queue", "", refBannerT, nil, many, 120)
	if EstimateTokensFromBytes(tight) > 120 {
		t.Fatalf("over budget: %d", EstimateTokensFromBytes(tight))
	}
	opened := strings.Count(tight, "--- [root/reference/n")
	if opened < len(many) && !strings.Contains(tight, "more concepts matched") && EstimateTokensFromBytes(tight+droppedNote([]string{"root/reference/n4"})) <= 120 {
		t.Errorf("dropped concepts must be named when the note fits:\n%s", tight)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies fence and heading parsing against CommonMark's rules where they decide a split.
// [Execution Surface]: Pure function
// [Assertions]: "```js" inside a fence does not close it; a 4-backtick fence needs 4 to close; a fence indented 4
// spaces is not a fence; "# C#" keeps its '#'; closing sequences are stripped; empty headings split.
// -----------------------------------------------------------------------------
func TestSplitSectionsFencesAndHeadings(t *testing.T) {
	content := "# A\n```\n```js\n# inside\n```\n## B ##\n````\n```\n# inside four\n````\n    ```\n### C#\n#\ntext"
	got := splitSections(content)
	var headings []string
	for _, s := range got {
		headings = append(headings, s.heading)
	}
	if want := []string{"A", "B", "C#", ""}; strings.Join(headings, "|") != strings.Join(want, "|") {
		t.Errorf("headings = %q, want %q", headings, want)
	}
	for in, want := range map[string]string{"# C#": "C#", "## B ##": "B", "#": "", "### x #y": "x #y", "## a#": "a#"} {
		if got := headingText(in); got != want {
			t.Errorf("headingText(%q) = %q, want %q", in, got, want)
		}
	}
}
