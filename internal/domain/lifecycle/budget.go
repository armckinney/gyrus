package lifecycle

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/armckinney/gyrus/internal/textrank"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

const (
	contextLayerHeader = "=== SUGGESTED CONTEXT LAYER ===\n\n"
	truncationMark     = "\n[...]"
)

func conceptBanner(doc gyrus.Concept) string {
	return fmt.Sprintf("--- [%s] %s (%s/%s) ---\n", doc.Name, doc.Title, doc.Category, doc.Type)
}

func omittedNote(doc gyrus.Concept, n int) string {
	return fmt.Sprintf("[%d more sections: gyrus get %s]\n\n", n, doc.Name)
}

func droppedNote(names []string) string {
	return fmt.Sprintf("[%d more concepts matched: gyrus get %s]\n\n", len(names), strings.Join(names, ", "))
}

// bytesPerToken is measured, not assumed: context layers of Markdown concepts (tables, ids, banners) injected into
// Claude Code prompts cost 2.6 bytes per token (median; interquartile range 2.52 to 2.79; 146 turns in 14 sessions,
// 2026-10-20). It sits at the low end on purpose: plain prose runs above it, so an estimate made with it errs high.
const bytesPerToken = 2.6

// EstimateTokensFromBytes estimates the token count of s from its length in bytes at bytesPerToken, rounded up:
// ceil(len(s)/2.6), computed in integers as (5*len(s)+12)/13. len(s) counts UTF-8 bytes, not characters. Estimates of
// pieces add up to at least the estimate of their concatenation, so a sum of piece estimates is a safe bound.
func EstimateTokensFromBytes(s string) int {
	return (5*len(s) + 12) / 13
}

// section is a run of a concept's content from one Markdown heading to the next heading of any level.
type section struct {
	heading string // the heading's text; empty for the preamble before the first heading
	text    string // the heading line and its body, trimmed
	order   int
}

// indent returns the width of line's leading spaces and whether it is indented as code (four or more, or a tab).
func indent(line string) (int, bool) {
	n := len(line) - len(strings.TrimLeft(line, " "))
	return n, n >= 4 || strings.HasPrefix(line[n:], "\t")
}

// isHeading reports whether line is an ATX heading (up to three leading spaces, one to six '#', then a space or end).
func isHeading(line string) bool {
	n, code := indent(line)
	if code {
		return false
	}
	t := line[n:]
	hashes := len(t) - len(strings.TrimLeft(t, "#"))
	return hashes >= 1 && hashes <= 6 && (hashes == len(t) || t[hashes] == ' ' || t[hashes] == '\t')
}

// headingText strips an ATX heading's opening hashes and its optional closing sequence ("## A #" is "A"); a '#' that
// is part of the text ("# C#") stays.
func headingText(line string) string {
	t := strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "#"))
	if closing := strings.TrimRight(t, "#"); closing != t && (closing == "" || strings.HasSuffix(closing, " ")) {
		t = strings.TrimSpace(closing)
	}
	return t
}

// fenceOpen reports the fence a line opens: three or more backticks or tildes, indented at most three spaces.
func fenceOpen(line string) (string, bool) {
	n, code := indent(line)
	if code {
		return "", false
	}
	t := line[n:]
	for _, c := range []string{"`", "~"} {
		if run := len(t) - len(strings.TrimLeft(t, c)); run >= 3 {
			return strings.Repeat(c, run), true
		}
	}
	return "", false
}

// fenceCloses reports whether line closes fence: the same character, at least as many, and nothing after but spaces.
func fenceCloses(line, fence string) bool {
	n, code := indent(line)
	if code {
		return false
	}
	t := strings.TrimRight(line[n:], " \t")
	return len(t) >= len(fence) && strings.Trim(t, fence[:1]) == ""
}

// splitSections splits Markdown content at every heading outside fenced code blocks.
func splitSections(content string) []section {
	var out []section
	var cur []string
	heading, fence := "", ""
	flush := func() {
		if text := strings.TrimSpace(strings.Join(cur, "\n")); text != "" {
			out = append(out, section{heading: heading, text: text, order: len(out)})
		}
		cur = nil
	}
	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		if fence != "" {
			if fenceCloses(line, fence) {
				fence = ""
			}
		} else if f, ok := fenceOpen(line); ok {
			fence = f
		} else if isHeading(line) {
			flush()
			heading = headingText(line)
		}
		cur = append(cur, line)
	}
	flush()
	return out
}

// truncateAtParagraphs keeps whole leading paragraphs of text that fit within maxTokens, marked as truncated.
func truncateAtParagraphs(text string, maxTokens int) (string, bool) {
	kept := ""
	for _, para := range strings.Split(text, "\n\n") {
		next := para
		if kept != "" {
			next = kept + "\n\n" + para
		}
		if EstimateTokensFromBytes(next+truncationMark+"\n\n") > maxTokens {
			break
		}
		kept = next
	}
	if kept == "" {
		return "", false
	}
	return kept + truncationMark, true
}

type candidate struct {
	doc   int
	sec   section
	score float64
}

// renderBudgeted builds the context layer from the concepts' sections that best match the prompt, within maxTokens.
//
// Sections are scored by BM25 against the prompt over all candidate sections, a heading's words counting twice; the
// score is relevance only, not relevance per token, so the fill below is greedy, not optimal. In priority order
// (project concepts first) each concept gets its best section, or its first when none matches, but no more than an
// even share of what is left, so one large section cannot crowd out the concepts after it; a section larger than its
// share is cut to its leading paragraphs. The remaining matching sections then fill the budget, best first, and a cut
// section is restored whole if it now fits. Sections keep document order; a concept with sections left out says how
// many, and concepts left out entirely are listed when the note fits. Nothing is cut inside a paragraph. A budget
// below the header's size yields "", and one that fits no paragraph yields the header alone.
func renderBudgeted(prompt, projectBanner, refBanner string, project, ref []gyrus.Concept, maxTokens int) string {
	docs := append(append([]gyrus.Concept{}, project...), ref...)
	isProject := func(i int) bool { return i < len(project) }

	terms := textrank.Unique(textrank.Tokenize(prompt))
	sections := make([][]section, len(docs))
	var tokens [][]string
	for i, doc := range docs {
		sections[i] = splitSections(doc.Content)
		for _, s := range sections[i] {
			tokens = append(tokens, append(textrank.Tokenize(s.heading), textrank.Tokenize(s.text)...))
		}
	}
	corpus := textrank.NewCorpus(tokens)
	scores := make([][]float64, len(docs))
	for i, k := 0, 0; i < len(docs); i++ {
		scores[i] = make([]float64, len(sections[i]))
		for j := range sections[i] {
			scores[i][j] = corpus.Score(k, terms)
			k++
		}
	}

	budget := maxTokens - EstimateTokensFromBytes(contextLayerHeader)
	if budget < 0 {
		return ""
	}
	bannerCost := make([]int, len(docs))
	for i, doc := range docs {
		bannerCost[i] = EstimateTokensFromBytes(conceptBanner(doc)) + EstimateTokensFromBytes(omittedNote(doc, len(sections[i])))
	}
	groupCost := map[bool]int{true: EstimateTokensFromBytes(projectBanner), false: EstimateTokensFromBytes(refBanner)}
	chosen := make([]map[int]string, len(docs))
	groupOpen := map[bool]bool{}
	openCost := func(i int) int {
		if chosen[i] != nil {
			return 0
		}
		cost := bannerCost[i]
		if !groupOpen[isProject(i)] {
			cost += groupCost[isProject(i)]
		}
		return cost
	}
	take := func(i int, s section, text string, cost int) {
		if chosen[i] == nil {
			chosen[i] = make(map[int]string)
			groupOpen[isProject(i)] = true
		}
		chosen[i][s.order] = text
		budget -= cost
	}

	var rest []candidate
	var cut []candidate
	left := 0
	for i := range docs {
		if len(sections[i]) > 0 {
			left++
		}
	}
	for i := range docs {
		if len(sections[i]) == 0 {
			continue
		}
		best := candidate{doc: i, sec: sections[i][0], score: -1}
		for j, s := range sections[i] {
			sc := scores[i][j]
			if sc > 0 {
				rest = append(rest, candidate{doc: i, sec: s, score: sc})
			}
			if sc > best.score {
				best = candidate{doc: i, sec: s, score: sc}
			}
		}
		share := budget / left
		left--
		open := openCost(i)
		if cost := open + EstimateTokensFromBytes(best.sec.text+"\n\n"); cost <= share {
			take(i, best.sec, best.sec.text, cost)
		} else if text, ok := truncateAtParagraphs(best.sec.text, share-open); ok {
			take(i, best.sec, text, open+EstimateTokensFromBytes(text+"\n\n"))
			cut = append(cut, best)
		}
	}

	slices.SortStableFunc(rest, func(a, b candidate) int { return cmp.Compare(b.score, a.score) })
	for _, c := range rest {
		if chosen[c.doc] != nil {
			if _, done := chosen[c.doc][c.sec.order]; done {
				continue
			}
		}
		if cost := openCost(c.doc) + EstimateTokensFromBytes(c.sec.text+"\n\n"); cost <= budget {
			take(c.doc, c.sec, c.sec.text, cost)
		}
	}
	for _, c := range cut {
		was := EstimateTokensFromBytes(chosen[c.doc][c.sec.order] + "\n\n")
		if delta := EstimateTokensFromBytes(c.sec.text+"\n\n") - was; delta <= budget {
			chosen[c.doc][c.sec.order] = c.sec.text
			budget -= delta
		}
	}

	var dropped []string
	for i := range docs {
		if len(sections[i]) > 0 && chosen[i] == nil {
			dropped = append(dropped, docs[i].Name)
		}
	}

	var sb strings.Builder
	sb.WriteString(contextLayerHeader)
	for _, group := range []struct {
		banner   string
		from, to int
	}{{projectBanner, 0, len(project)}, {refBanner, len(project), len(docs)}} {
		bannerDone := false
		for i := group.from; i < group.to; i++ {
			if chosen[i] == nil {
				continue
			}
			if !bannerDone {
				sb.WriteString(group.banner)
				bannerDone = true
			}
			sb.WriteString(conceptBanner(docs[i]))
			for _, s := range sections[i] {
				if text, ok := chosen[i][s.order]; ok {
					sb.WriteString(text)
					sb.WriteString("\n\n")
				}
			}
			if omitted := len(sections[i]) - len(chosen[i]); omitted > 0 {
				sb.WriteString(omittedNote(docs[i], omitted))
			}
		}
	}
	if len(dropped) > 0 {
		if note := droppedNote(dropped); EstimateTokensFromBytes(note) <= budget {
			sb.WriteString(note)
		}
	}
	return sb.String()
}
