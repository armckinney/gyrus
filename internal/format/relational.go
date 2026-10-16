package format

import (
	"encoding/json"
	"fmt"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

// RelationalFormat implements StorageFormat for JSON relational envelopes.
type RelationalFormat struct{}

// NewRelationalFormat returns a new RelationalFormat serializer instance.
func NewRelationalFormat() *RelationalFormat {
	return &RelationalFormat{}
}

// Serialize converts a gyrus.Concept into JSON relational payload bytes.
func (f *RelationalFormat) Serialize(doc *gyrus.Concept) ([]byte, error) {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize document to JSON: %w", err)
	}
	return data, nil
}

// Deserialize parses JSON relational payload bytes into a gyrus.Concept.
func (f *RelationalFormat) Deserialize(data []byte) (*gyrus.Concept, error) {
	var doc gyrus.Concept
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse relational JSON envelope: %w", err)
	}
	if doc.Format == "" {
		doc.Format = "json"
	}
	return &doc, nil
}
