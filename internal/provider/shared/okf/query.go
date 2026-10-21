// Package okf implements shared ephemeral OKF snapshots and BM25 keyword ranking.
package okf

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/internal/textrank"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

const (
	exactIDBoost = 1000
	// projectBoost and referenceBoost multiply the BM25 score of a concept that matched the query: scope reorders
	// matches but never makes a concept match.
	projectBoost   = 1.5
	referenceBoost = 1.2
	// identityRepeat and tagRepeat weight the name, title and tags by repeating their terms in the indexed text.
	identityRepeat = 3
	tagRepeat      = 2
	// Browsing with an empty query orders project concepts, then reference concepts, then the rest.
	browseProject   = 60
	browseReference = 20
)

func Scanner(store gyrus.ConceptStore) (gyrus.ConceptScanner, error) {
	s, ok := store.(gyrus.ConceptScanner)
	if !ok {
		return nil, fmt.Errorf("OKF providers require scanner-capable storage")
	}
	return s, nil
}

func Snapshot(ctx context.Context, scanner gyrus.ConceptScanner) ([]gyrus.Concept, error) {
	docs := make([]gyrus.Concept, 0)
	report, err := scanner.Scan(ctx, func(doc gyrus.Concept) error { docs = append(docs, doc); return nil })
	if err != nil {
		return nil, err
	}
	if len(report.Errors) > 0 {
		return nil, &okf.ValidationError{Field: "collection", Message: strings.Join(report.Errors, "; ")}
	}
	return docs, nil
}

func matches(doc gyrus.Concept, f gyrus.SearchFilter) bool {
	if f.Category != "" && doc.Category != f.Category || f.Type != "" && doc.Type != f.Type || f.Status != "" && doc.Status != f.Status || f.Tenant != "" && doc.Tenant != f.Tenant || f.Tag != "" && !slices.Contains(doc.Tags, f.Tag) {
		return false
	}
	switch f.Scope {
	case "all":
		return true
	case "reference":
		return doc.Scope == "reference"
	case "project":
		return doc.Scope == "project" && (f.Project == "" || doc.Project == f.Project)
	default:
		return doc.Scope == "reference" || f.Project != "" && doc.Project == f.Project
	}
}

// indexed is a concept's searchable terms, weighted for BM25, and its short identity fields for the match reason;
// a term in none of those must be in the body, which the corpus answers.
type indexed struct {
	tokens          []string
	id, title, tags []string
}

func index(doc gyrus.Concept, o textrank.Options) indexed {
	id := textrank.TokenizeWith(doc.Name, o)
	title := textrank.TokenizeWith(doc.Title, o)
	tags := textrank.TokenizeWith(strings.Join(doc.Tags, " "), o)
	body := textrank.TokenizeWith(doc.Content, o)
	var all []string
	for i := 0; i < identityRepeat; i++ {
		all = append(append(all, id...), title...)
	}
	for i := 0; i < tagRepeat; i++ {
		all = append(all, tags...)
	}
	all = append(all, body...)
	return indexed{all, id, title, tags}
}

func browse(doc gyrus.Concept, f gyrus.SearchFilter) (float64, string) {
	if f.Project != "" && doc.Project == f.Project {
		return browseProject, "Project Concept"
	}
	if doc.Scope == "reference" {
		return browseReference, "Reference Concept"
	}
	return 0, "All Concepts"
}

// score is BM25 times the scope boost, plus exactIDBoost when the query is the concept's name; ok is false when no
// query term occurs in the concept.
func score(doc gyrus.Concept, ix indexed, corpus *textrank.Corpus, i int, raw string, terms []string, f gyrus.SearchFilter) (float64, string, bool) {
	bm25 := corpus.Score(i, terms)
	exact := strings.ToLower(doc.Name) == raw
	if bm25 == 0 && !exact {
		return 0, "", false
	}
	bestField, reason := 0, ""
	for _, term := range terms {
		field, r := 0, ""
		switch {
		case slices.Contains(ix.id, term):
			field, r = 4, "Concept ID Match"
		case slices.Contains(ix.title, term):
			field, r = 3, "Title Match"
		case slices.Contains(ix.tags, term):
			field, r = 2, "Tag Match"
		case corpus.TermFreq(i, term) > 0:
			field, r = 1, "Content Match"
		}
		if field > bestField {
			bestField, reason = field, r
		}
	}
	value := bm25
	if f.Project != "" && doc.Project == f.Project {
		value *= projectBoost
		reason = "Project Match"
	} else if doc.Scope == "reference" {
		value *= referenceBoost
	}
	if exact {
		value += exactIDBoost
		reason = "Exact ID Match"
	}
	return value, reason, true
}

func Search(ctx context.Context, scanner gyrus.ConceptScanner, q gyrus.SearchQuery) ([]gyrus.SearchResult, error) {
	switch q.Filter.Scope {
	case "", "all", "project", "reference":
	default:
		return nil, &okf.ValidationError{Field: "scope", Message: "must be project, reference, or all"}
	}
	docs, err := Snapshot(ctx, scanner)
	if err != nil {
		return nil, err
	}
	raw := strings.ToLower(strings.TrimSpace(q.Query))
	candidates := make([]gyrus.Concept, 0, len(docs))
	for _, doc := range docs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if matches(doc, q.Filter) {
			candidates = append(candidates, doc)
		}
	}
	results := make([]gyrus.SearchResult, 0)
	if raw == "" {
		for _, doc := range candidates {
			value, reason := browse(doc, q.Filter)
			results = append(results, gyrus.SearchResult{Concept: doc, Score: value, MatchReason: reason})
		}
	} else {
		// Term statistics come from the concepts the filter admits, so rarity is judged among what can be returned.
		opts, terms := textrank.ForQuery(raw)
		indexes := make([]indexed, len(candidates))
		tokens := make([][]string, len(candidates))
		for i, doc := range candidates {
			indexes[i] = index(doc, opts)
			tokens[i] = indexes[i].tokens
		}
		corpus := textrank.NewCorpus(tokens)
		for i, doc := range candidates {
			value, reason, ok := score(doc, indexes[i], corpus, i, raw, terms, q.Filter)
			if ok {
				results = append(results, gyrus.SearchResult{Concept: doc, Score: value, MatchReason: reason})
			}
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Concept.Name < results[j].Concept.Name
	})
	q.Query = raw
	limit := q.Limit()
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}
