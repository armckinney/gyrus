// Package okf persists graph mutations in source concept frontmatter.
package okf

import (
	"context"
	"slices"
	"sort"

	domain "github.com/armckinney/gyrus/internal/domain/okf"
	sharedokf "github.com/armckinney/gyrus/internal/provider/shared/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

type GraphStore struct {
	store   gyrus.ConceptStore
	scanner gyrus.ConceptScanner
}

func NewGraphStore(store gyrus.ConceptStore) (*GraphStore, error) {
	scanner, err := sharedokf.Scanner(store)
	if err != nil {
		return nil, err
	}
	return &GraphStore{store, scanner}, nil
}

// StoresLinksInConcepts marks a graph whose mutations need secondary metadata refresh.
func (*GraphStore) StoresLinksInConcepts() bool { return true }

type edgeKey struct {
	from, to string
	rel      gyrus.RelationshipType
}

func key(e gyrus.ConceptEdge) edgeKey {
	return edgeKey{e.FromConceptID, e.ToConceptID, e.RelationshipType}
}
func validateEdge(e gyrus.ConceptEdge) error {
	return domain.ValidateEdge(e)
}

func (g *GraphStore) UpsertEdges(ctx context.Context, edges []gyrus.ConceptEdge) error {
	groups := make(map[string][]gyrus.ConceptEdge)
	for _, edge := range edges {
		if err := validateEdge(edge); err != nil {
			return err
		}
		groups[edge.FromConceptID] = append(groups[edge.FromConceptID], edge)
	}
	ids := make([]string, 0, len(groups))
	for id := range groups {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	docs := make(map[string]gyrus.Concept)
	for _, id := range ids {
		doc, err := g.store.Get(ctx, id)
		if err != nil {
			return err
		}
		if err := domain.Validate(&doc); err != nil {
			return err
		}
		docs[id] = doc
	}
	applied := make([]string, 0, len(ids))
	for _, id := range ids {
		doc := docs[id]
		present := make(map[edgeKey]bool)
		for _, edge := range gyrus.ConceptEdges(doc) {
			present[key(edge)] = true
		}
		links := slices.Clone(doc.Links)
		changed := false
		for _, edge := range groups[id] {
			if present[key(edge)] {
				continue
			}
			present[key(edge)] = true
			links = append(links, gyrus.ConceptLink{Concept: edge.ToConceptID, Rel: edge.RelationshipType})
			changed = true
		}
		if !changed {
			continue
		}
		if _, err := g.store.Update(ctx, id, gyrus.ConceptPatch{Links: &links}, doc.Version); err != nil {
			return &gyrus.PartialGraphMutationError{AppliedSourceIDs: applied, FailedSourceID: id, Err: err}
		}
		applied = append(applied, id)
	}
	return nil
}
func (g *GraphStore) DeleteEdges(ctx context.Context, from, to string, rel gyrus.RelationshipType) error {
	if err := validateEdge(gyrus.ConceptEdge{FromConceptID: from, ToConceptID: to, RelationshipType: rel}); err != nil {
		return err
	}
	doc, err := g.store.Get(ctx, from)
	if err != nil {
		return err
	}
	links := slices.DeleteFunc(slices.Clone(doc.Links), func(l gyrus.ConceptLink) bool { return l.Concept == to && l.Rel == rel })
	if len(links) == len(doc.Links) {
		return nil
	}
	_, err = g.store.Update(ctx, from, gyrus.ConceptPatch{Links: &links}, doc.Version)
	return err
}
func (g *GraphStore) snapshot(ctx context.Context) ([]gyrus.ConceptEdge, error) {
	docs, err := sharedokf.Snapshot(ctx, g.scanner)
	if err != nil {
		return nil, err
	}
	edges := make([]gyrus.ConceptEdge, 0)
	for _, doc := range docs {
		edges = append(edges, gyrus.ConceptEdges(doc)...)
	}
	gyrus.SortEdges(edges)
	return edges, nil
}
func (g *GraphStore) Neighbors(ctx context.Context, id string, f gyrus.EdgeFilter) ([]gyrus.ConceptEdge, error) {
	if err := domain.ValidateName(id); err != nil {
		return nil, err
	}
	if f.Direction == "" {
		f.Direction = "both"
	}
	if f.Direction != "both" && f.Direction != "incoming" && f.Direction != "outgoing" {
		return nil, &domain.ValidationError{Field: "direction", Message: "must be incoming, outgoing, or both"}
	}
	if f.RelationshipType != "" && !gyrus.ValidRelationship(f.RelationshipType) {
		return nil, &domain.ValidationError{Field: "rel_type", Message: "invalid relationship type"}
	}
	edges, err := g.snapshot(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]gyrus.ConceptEdge, 0)
	for _, edge := range edges {
		if f.RelationshipType != "" && edge.RelationshipType != f.RelationshipType {
			continue
		}
		if edge.FromConceptID == id && f.Direction != "incoming" || edge.ToConceptID == id && f.Direction != "outgoing" {
			result = append(result, edge)
		}
	}
	return result, nil
}
func (g *GraphStore) Traverse(ctx context.Context, q gyrus.GraphQuery) ([]gyrus.GraphPath, error) {
	if err := domain.ValidateName(q.StartID); err != nil {
		return nil, err
	}
	if q.MaxDepth <= 0 {
		q.MaxDepth = 1
	}
	edges, err := g.snapshot(ctx)
	if err != nil {
		return nil, err
	}
	adj := make(map[string][]gyrus.ConceptEdge)
	for _, edge := range edges {
		adj[edge.FromConceptID] = append(adj[edge.FromConceptID], edge)
	}
	seen := map[string]bool{q.StartID: true}
	visited := make(map[edgeKey]bool)
	level := []string{q.StartID}
	path := gyrus.GraphPath{Edges: make([]gyrus.ConceptEdge, 0)}
	for depth := 0; depth < q.MaxDepth && len(level) > 0; depth++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var next []string
		for _, id := range level {
			for _, edge := range adj[id] {
				if !visited[key(edge)] {
					path.Edges = append(path.Edges, edge)
					visited[key(edge)] = true
				}
				if !seen[edge.ToConceptID] {
					seen[edge.ToConceptID] = true
					next = append(next, edge.ToConceptID)
				}
			}
		}
		level = next
	}
	for id := range seen {
		path.Nodes = append(path.Nodes, id)
	}
	sort.Strings(path.Nodes)
	gyrus.SortEdges(path.Edges)
	return []gyrus.GraphPath{path}, nil
}

var _ gyrus.GraphStore = (*GraphStore)(nil)
