// Package okf implements shared ephemeral OKF snapshots and keyword matching.
package okf

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"unicode"

	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

const (
	exactIDBoost   = 1000
	projectBoost   = 60
	referenceBoost = 20
	identityWeight = 10
	tagWeight      = 5
	bodyWeight     = 1
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

func score(doc gyrus.Concept, raw string, terms []string, f gyrus.SearchFilter) (float64, string, bool) {
	value := 0.0
	project := f.Project != "" && doc.Project == f.Project
	if project {
		value = projectBoost
	} else if doc.Scope == "reference" {
		value = referenceBoost
	}
	if raw == "" {
		reason := "All Concepts"
		if project {
			reason = "Project Concept"
		} else if doc.Scope == "reference" {
			reason = "Reference Concept"
		}
		return value, reason, true
	}
	id, title, body := strings.ToLower(doc.Name), strings.ToLower(doc.Title), strings.ToLower(doc.Content)
	var total float64
	bestReason := ""
	bestField := 0
	for _, term := range terms {
		weight, field, reason := 0, 0, ""
		switch {
		case strings.Contains(id, term):
			weight, field, reason = identityWeight, 4, "Concept ID Match"
		case strings.Contains(title, term):
			weight, field, reason = identityWeight, 3, "Title Match"
		default:
			for _, tag := range doc.Tags {
				if strings.Contains(strings.ToLower(tag), term) {
					weight, field, reason = tagWeight, 2, "Tag Match"
					break
				}
			}
			if weight == 0 && strings.Contains(body, term) {
				weight, field, reason = bodyWeight, 1, "Content Match"
			}
		}
		total += float64(weight)
		if field > bestField {
			bestField, bestReason = field, reason
		}
	}
	if total == 0 {
		return 0, "", false
	}
	value += total / float64(len(terms))
	if project {
		bestReason = "Project Match"
	}
	if id == raw {
		value += exactIDBoost
		bestReason = "Exact ID Match"
	}
	return value, bestReason, true
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
	parts := strings.FieldsFunc(raw, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' })
	terms := make([]string, 0, len(parts))
	seen := make(map[string]bool)
	for _, term := range parts {
		if !seen[term] {
			terms = append(terms, term)
			seen[term] = true
		}
	}
	results := make([]gyrus.SearchResult, 0)
	for _, doc := range docs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !matches(doc, q.Filter) {
			continue
		}
		value, reason, ok := score(doc, raw, terms, q.Filter)
		if ok {
			results = append(results, gyrus.SearchResult{Concept: doc, Score: value, MatchReason: reason})
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
