// Package textrank tokenizes text into search terms and ranks token lists against a query with Okapi BM25.
package textrank

import (
	"math"
	"strings"
	"unicode"
)

const (
	k1 = 1.2
	b  = 0.75
)

var stopwords = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true, "at": true, "be": true, "by": true, "can": true,
	"do": true, "does": true, "for": true, "from": true, "how": true, "i": true, "if": true, "in": true, "is": true,
	"it": true, "its": true, "me": true, "my": true, "of": true, "on": true, "or": true, "should": true, "so": true,
	"that": true, "the": true, "this": true, "to": true, "we": true, "what": true, "when": true, "where": true,
	"which": true, "who": true, "why": true, "will": true, "with": true, "you": true, "your": true,
}

// stem folds a plural onto its singular ("tasks" -> "task"); it is deliberately minimal.
func stem(t string) string {
	if len(t) > 3 && strings.HasSuffix(t, "s") && !strings.HasSuffix(t, "ss") && !strings.HasSuffix(t, "us") {
		return t[:len(t)-1]
	}
	return t
}

// Tokenize lowercases text and splits it into terms: letters and digits, with stopwords dropped. A hyphenated or
// underscored word yields itself and its parts, so "agent-ops" matches both "agent-ops" and "agent ops".
func Tokenize(text string) []string {
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_'
	})
	var out []string
	add := func(w string) {
		w = strings.Trim(w, "-_")
		if w != "" && !stopwords[w] {
			out = append(out, stem(w))
		}
	}
	for _, w := range words {
		add(w)
		if strings.ContainsAny(w, "-_") {
			for _, part := range strings.FieldsFunc(w, func(r rune) bool { return r == '-' || r == '_' }) {
				add(part)
			}
		}
	}
	return out
}

// Unique returns the distinct terms of a token list in first-seen order.
func Unique(tokens []string) []string {
	seen := make(map[string]bool, len(tokens))
	var out []string
	for _, t := range tokens {
		if !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}

// Corpus holds term statistics for BM25 over a fixed set of token lists.
type Corpus struct {
	tf     []map[string]int
	length []int
	df     map[string]int
	avg    float64
}

// NewCorpus indexes each token list as one document.
func NewCorpus(docs [][]string) *Corpus {
	c := &Corpus{tf: make([]map[string]int, len(docs)), length: make([]int, len(docs)), df: make(map[string]int)}
	total := 0
	for i, tokens := range docs {
		c.tf[i] = make(map[string]int)
		for _, t := range tokens {
			c.tf[i][t]++
		}
		for t := range c.tf[i] {
			c.df[t]++
		}
		c.length[i] = len(tokens)
		total += len(tokens)
	}
	if len(docs) > 0 {
		c.avg = float64(total) / float64(len(docs))
	}
	return c
}

// Score is the BM25 score of document i for the distinct query terms; zero when no term occurs in it.
func (c *Corpus) Score(i int, terms []string) float64 {
	if c.avg == 0 {
		return 0
	}
	n := float64(len(c.tf))
	norm := k1 * (1 - b + b*float64(c.length[i])/c.avg)
	score := 0.0
	for _, t := range terms {
		f := float64(c.tf[i][t])
		if f == 0 {
			continue
		}
		df := float64(c.df[t])
		idf := math.Log(1 + (n-df+0.5)/(df+0.5))
		score += idf * f * (k1 + 1) / (f + norm)
	}
	return score
}
