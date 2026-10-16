package gyrus

import "sort"

func ValidRelationship(rel RelationshipType) bool {
	switch rel {
	case RelDependsOn, RelSupersedes, RelImplements, RelMitigates:
		return true
	}
	return false
}

// ConceptEdges returns the deduplicated source-owned relationships.
func ConceptEdges(doc Concept) []ConceptEdge {
	type key struct {
		target string
		rel    RelationshipType
	}
	edges := make(map[key]ConceptEdge)
	for _, link := range doc.Links {
		k := key{link.Concept, link.Rel}
		edges[k] = ConceptEdge{FromConceptID: doc.Name, ToConceptID: link.Concept, RelationshipType: link.Rel}
	}
	result := make([]ConceptEdge, 0, len(edges))
	for _, edge := range edges {
		result = append(result, edge)
	}
	SortEdges(result)
	return result
}

func SortEdges(edges []ConceptEdge) {
	sort.Slice(edges, func(i, j int) bool {
		a, b := edges[i], edges[j]
		if a.FromConceptID != b.FromConceptID {
			return a.FromConceptID < b.FromConceptID
		}
		if a.ToConceptID != b.ToConceptID {
			return a.ToConceptID < b.ToConceptID
		}
		return a.RelationshipType < b.RelationshipType
	})
}
