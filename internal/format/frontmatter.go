package format

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

var frontmatterOrder = []string{
	"title", "id", "type", "format", "category", "version", "status",
	"last_updated", "last_modified_by", "tenant", "tags", "links",
}

// MarshalFrontmatter applies the canonical presentation order to YAML metadata.
func MarshalFrontmatter(value any) ([]byte, error) {
	var node yaml.Node
	if err := node.Encode(value); err != nil {
		return nil, err
	}
	return marshalOrderedFrontmatter(&node)
}

// OrderFrontmatter reorders existing YAML while retaining values and comments.
func OrderFrontmatter(data []byte) ([]byte, error) {
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, err
	}
	return marshalOrderedFrontmatter(&node)
}

func marshalOrderedFrontmatter(node *yaml.Node) ([]byte, error) {
	mapping := node
	if mapping.Kind == yaml.DocumentNode && len(mapping.Content) == 1 {
		mapping = mapping.Content[0]
	}
	if mapping.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("frontmatter must be a YAML mapping")
	}
	ranks := make(map[string]int, len(frontmatterOrder))
	for i, key := range frontmatterOrder {
		ranks[key] = i
	}
	type field struct {
		key, value *yaml.Node
		rank       int
	}
	fields := make([]field, 0, len(mapping.Content)/2)
	for i := 0; i < len(mapping.Content); i += 2 {
		rank, known := ranks[mapping.Content[i].Value]
		if !known {
			rank = len(frontmatterOrder)
		}
		fields = append(fields, field{mapping.Content[i], mapping.Content[i+1], rank})
	}
	sort.SliceStable(fields, func(i, j int) bool { return fields[i].rank < fields[j].rank })
	mapping.Content = mapping.Content[:0]
	for _, field := range fields {
		mapping.Content = append(mapping.Content, field.key, field.value)
	}
	return yaml.Marshal(node)
}
