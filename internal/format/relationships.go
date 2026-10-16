package format

import "fmt"

func validateRelationships(fields map[string]any) error {
	if _, exists := fields["dependencies"]; exists {
		return fmt.Errorf("obsolete dependencies field: convert to links with concept and rel: depends_on")
	}
	if raw, exists := fields["links"]; exists && raw != nil {
		links, ok := raw.([]any)
		if !ok {
			return fmt.Errorf("links must be a list of {concept, rel} mappings")
		}
		for _, rawLink := range links {
			link, ok := rawLink.(map[string]any)
			if !ok {
				return fmt.Errorf("each link must be a {concept, rel} mapping")
			}
			for key := range link {
				if key != "concept" && key != "rel" {
					return fmt.Errorf("unsupported links.%s: use only concept and rel; convert obsolete relationship metadata first", key)
				}
			}
			for _, key := range []string{"concept", "rel"} {
				if value, ok := link[key].(string); !ok || value == "" {
					return fmt.Errorf("links.%s is required and must be a nonempty string", key)
				}
			}
		}
	}
	return nil
}
