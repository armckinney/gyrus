package format_test

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/armckinney/gyrus/internal/format"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"gopkg.in/yaml.v3"
)

func TestCanonicalFrontmatterOrder(t *testing.T) {
	doc := gyrus.Concept{ID: "adr-example", Title: "Example", Type: gyrus.TypeADR, Format: "markdown", Category: gyrus.CategoryArchitecture, Version: 2, Status: "accepted", LastUpdated: time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), LastModifiedBy: "author", Tenant: "root", Tags: []string{"example"}, Links: []gyrus.ConceptLink{{Concept: "root/reference/spec-example", Rel: gyrus.RelDependsOn}}, Immutable: true}
	data, err := format.NewOKFFormat().Serialize(&doc)
	if err != nil {
		t.Fatal(err)
	}
	parts := bytes.SplitN(data[4:], []byte("\n---"), 2)
	var node yaml.Node
	if err := yaml.Unmarshal(parts[0], &node); err != nil {
		t.Fatal(err)
	}
	var keys []string
	for i := 0; i < len(node.Content[0].Content); i += 2 {
		keys = append(keys, node.Content[0].Content[i].Value)
	}
	want := []string{"title", "id", "type", "format", "category", "version", "status", "last_updated", "last_modified_by", "tenant", "tags", "links", "immutable"}
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("order: got %v, want %v", keys, want)
	}
	var fields map[string]any
	if err := yaml.Unmarshal(parts[0], &fields); err != nil {
		t.Fatal(err)
	}
	fromMap, err := format.MarshalFrontmatter(fields)
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(fromMap, &node); err != nil {
		t.Fatal(err)
	}
	keys = nil
	for i := 0; i < len(node.Content[0].Content); i += 2 {
		keys = append(keys, node.Content[0].Content[i].Value)
	}
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("map serialization order: got %v, want %v", keys, want)
	}
}

func TestFrontmatterOrderingPreservesMetadataAndIsIdempotent(t *testing.T) {
	before := []byte("custom: value\nlinks: []\nimmutable: true\n# Title comment\ntitle: Example\nid: example\nversion: 7\n")
	after, err := format.OrderFrontmatter(before)
	if err != nil {
		t.Fatal(err)
	}
	var oldFields, newFields map[string]any
	if err := yaml.Unmarshal(before, &oldFields); err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(after, &newFields); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(oldFields, newFields) {
		t.Fatal("ordering changed metadata values")
	}
	if !bytes.Contains(after, []byte("# Title comment")) {
		t.Fatal("ordering lost comment")
	}
	repeat, err := format.OrderFrontmatter(after)
	if err != nil || !bytes.Equal(after, repeat) {
		t.Fatalf("not idempotent: %v", err)
	}
	if _, err := format.OrderFrontmatter([]byte("- value\n")); err == nil {
		t.Fatal("non-mapping accepted")
	}
}
