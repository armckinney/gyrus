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

// Options switches the tokenizer's normalisations. The zero value drops stopwords.
type Options struct {
	KeepStopwords bool // keep stopwords as terms
}

// Tokenize is TokenizeWith the default Options.
func Tokenize(text string) []string { return TokenizeWith(text, Options{}) }

// TokenizeWith lowercases text and splits it into terms: letters and digits, by default with stopwords dropped.
// Words are not stemmed: folding plurals measured worse on a real store. A hyphenated or underscored word yields
// itself and its parts, so "agent-ops" matches both "agent-ops" and "agent ops".
func TokenizeWith(text string, o Options) []string {
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_'
	})
	var out []string
	add := func(w string) {
		w = strings.Trim(w, "-_")
		if w != "" && (o.KeepStopwords || !stopwords[w]) {
			out = append(out, w)
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

// ForQuery returns the Options to tokenize a query and the text it is scored against, and the query's distinct
// terms. A query made only of stopwords ("what is this?") keeps them on both sides, so it still matches something
// instead of nothing.
func ForQuery(query string) (Options, []string) {
	if terms := Unique(Tokenize(query)); len(terms) > 0 {
		return Options{}, terms
	}
	o := Options{KeepStopwords: true}
	return o, Unique(TokenizeWith(query, o))
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

// TermFreq is how many times term occurs in document i.
func (c *Corpus) TermFreq(i int, term string) int { return c.tf[i][term] }

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
