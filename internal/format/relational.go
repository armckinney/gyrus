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

// Serialize converts a gyrus.Document into JSON relational payload bytes.
func (f *RelationalFormat) Serialize(doc *gyrus.Document) ([]byte, error) {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize document to JSON: %w", err)
	}
	return data, nil
}

// Deserialize parses JSON relational payload bytes into a gyrus.Document.
func (f *RelationalFormat) Deserialize(data []byte) (*gyrus.Document, error) {
	var doc gyrus.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse relational JSON envelope: %w", err)
	}
	if doc.Format == "" {
		doc.Format = "json"
	}
	return &doc, nil
}
