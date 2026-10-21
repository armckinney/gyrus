package lifecycle

import (
	"fmt"
	"sort"
	"strings"

	"github.com/armckinney/gyrus/internal/textrank"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

const (
	contextLayerHeader = "=== SUGGESTED CONTEXT LAYER ===\n\n"
	truncationMark     = "\n[…]"
)

func conceptBanner(doc gyrus.Concept) string {
	return fmt.Sprintf("--- [%s] %s (%s/%s) ---\n", doc.Name, doc.Title, doc.Category, doc.Type)
}

func omittedNote(doc gyrus.Concept, n int) string {
	return fmt.Sprintf("[%d more sections: gyrus get %s]\n\n", n, doc.Name)
}

// EstimateTokensFromBytes conservatively estimates the token count of s from its length in bytes, at 3.5 bytes per
// token, rounded up. len(s) counts UTF-8 bytes, not characters, so non-ASCII text estimates high, never low. Estimates
// of pieces add up to at least the estimate of their concatenation, so a sum of piece estimates is a safe bound.
func EstimateTokensFromBytes(s string) int {
	return (2*len(s) + 6) / 7
}

// section is a run of a concept's content from one Markdown heading to the next heading of any level.
type section struct {
	heading string // the heading's text; empty for the preamble before the first heading
	text    string // the heading line and its body, trimmed
	order   int
}

// isHeading reports whether line is an ATX heading (up to three leading spaces, one to six '#', then a space).
func isHeading(line string) bool {
	if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
		return false
	}
	t := strings.TrimLeft(line, " ")
	n := 0
	for n < len(t) && t[n] == '#' {
		n++
	}
	return n >= 1 && n <= 6 && (n == len(t) || t[n] == ' ' || t[n] == '\t')
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
		trimmed := strings.TrimSpace(line)
		switch {
		case fence != "":
			if strings.HasPrefix(trimmed, fence) {
				fence = ""
			}
		case strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~"):
			fence = trimmed[:3]
		case isHeading(line):
			flush()
			heading = strings.TrimSpace(strings.Trim(trimmed, "#"))
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
// Sections are scored by BM25 against the prompt, over all candidate sections (a heading's words count twice).
// Every concept first gets its best section (its first section when
// none matches) in priority order, project concepts first; the remaining matching sections then fill what is left of
// the budget, best first. A concept's sections keep their document order, and a concept whose best section does not
// fit gets its leading paragraphs instead. Nothing is cut inside a paragraph.
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
	chosen := make([]map[int]string, len(docs))
	groupOpen := map[bool]bool{}
	openCost := func(i int) int {
		if chosen[i] != nil {
			return 0
		}
		cost := EstimateTokensFromBytes(conceptBanner(docs[i])) + EstimateTokensFromBytes(omittedNote(docs[i], len(sections[i])))
		if !groupOpen[isProject(i)] {
			if isProject(i) {
				cost += EstimateTokensFromBytes(projectBanner)
			} else {
				cost += EstimateTokensFromBytes(refBanner)
			}
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
		open := openCost(i)
		if cost := open + EstimateTokensFromBytes(best.sec.text+"\n\n"); cost <= budget {
			take(i, best.sec, best.sec.text, cost)
		} else if text, ok := truncateAtParagraphs(best.sec.text, budget-open); ok {
			take(i, best.sec, text, open+EstimateTokensFromBytes(text+"\n\n"))
		}
	}

	sort.SliceStable(rest, func(a, b int) bool { return rest[a].score > rest[b].score })
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
	return sb.String()
}
