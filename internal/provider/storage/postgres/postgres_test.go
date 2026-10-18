package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/pashagolub/pgxmock/v4"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/internal/mcp/tools"
	mcpapi "github.com/mark3labs/mcp-go/mcp"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

func TestPostgresVersionConflictIsMCPToolError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()
	store := &Store{pool: mock}
	doc := gyrus.Concept{ID: "shared", Tenant: "root", Title: "Shared", Type: gyrus.TypeGuide, Category: gyrus.CategoryTechnical, Status: "draft", Version: 2, Scope: "reference"}
	metadata, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT frontmatter, content FROM concepts WHERE id = \\$1$").WithArgs("root/reference/shared").WillReturnRows(pgxmock.NewRows([]string{"frontmatter", "content"}).AddRow(metadata, "winner"))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT frontmatter, content FROM concepts WHERE id = \\$1 FOR UPDATE").WithArgs("root/reference/shared").WillReturnRows(pgxmock.NewRows([]string{"frontmatter", "content"}).AddRow(metadata, "winner"))
	mock.ExpectRollback()
	engine := lifecycle.NewEngine(store, nil, nil, nil, "")
	req := mcpapi.CallToolRequest{}
	req.Params.Arguments = map[string]any{"id": "root/reference/shared", "title": "loser", "expected_version": 1}
	result, err := tools.NewHandler(engine).HandleUpdate(context.Background(), req)
	if err != nil || !result.IsError {
		t.Fatalf("expected MCP tool error: %+v %v", result, err)
	}
	text := result.Content[0].(mcpapi.TextContent).Text
	if !strings.Contains(text, "concurrency conflict") || !strings.Contains(text, "root/reference/shared") {
		t.Fatalf("conflict details missing: %s", text)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStoreVersionConflictRollsBackWithoutWrites(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()
	store := &Store{pool: mock}
	doc := gyrus.Concept{ID: "shared", Tenant: "root", Version: 2, Scope: "reference"}
	metadata, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT frontmatter, content FROM concepts WHERE id = \\$1 FOR UPDATE").WithArgs("root/reference/shared").WillReturnRows(pgxmock.NewRows([]string{"frontmatter", "content"}).AddRow(metadata, "winner"))
	mock.ExpectRollback()
	title := "loser"
	_, err = store.Update(context.Background(), "root/reference/shared", gyrus.ConceptPatch{Title: &title}, 1)
	var conflict *gyrus.ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("expected typed conflict, got %T %v", err, err)
	}
	if conflict.ConceptID != "root/reference/shared" || conflict.Operation != "update" || !strings.Contains(conflict.Error(), "expected version 1, got 2") {
		t.Fatalf("incomplete conflict: %+v", conflict)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies that PostgreSQL Store creates a concept and appends an entry to concepts_history.
// [Execution Surface]: In-Memory PostgreSQL Mock Pool (pgxmock)
// [Assertions]: INSERT queries for concepts and concepts_history execute in a transaction.
// -----------------------------------------------------------------------------
func TestStore_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	store := &Store{pool: mock}

	doc := gyrus.Concept{
		ID:      "doc-1",
		Title:   "Test Doc",
		Content: "Hello World", Tenant: "root", Name: "root/reference/doc-1",
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO concepts").WithArgs("root/reference/doc-1", pgxmock.AnyArg(), "Hello World").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("INSERT INTO concepts_history").WithArgs("root/reference/doc-1", 1, pgxmock.AnyArg(), "Hello World", pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	ref, err := store.Create(context.Background(), doc)
	if err != nil {
		t.Errorf("error was not expected while creating doc: %s", err)
	}

	if ref.ID != "doc-1" {
		t.Errorf("expected id 'doc-1', got '%s'", ref.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies that PostgreSQL Store fetches concept frontmatter JSON and content.
// [Execution Surface]: In-Memory PostgreSQL Mock Pool (pgxmock)
// [Assertions]: SELECT query returns frontmatter and content fields without error.
// -----------------------------------------------------------------------------
func TestStore_Get(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	store := &Store{pool: mock}

	doc := gyrus.Concept{
		ID:    "doc-1",
		Title: "Test Doc", Tenant: "root", Name: "root/reference/doc-1",
	}
	fmBytes, _ := json.Marshal(doc)

	rows := pgxmock.NewRows([]string{"frontmatter", "content"}).
		AddRow(fmBytes, "Hello World")

	mock.ExpectQuery("SELECT frontmatter, content FROM concepts WHERE id = \\$1").
		WithArgs("root/reference/doc-1").
		WillReturnRows(rows)

	res, err := store.Get(context.Background(), "root/reference/doc-1")
	if err != nil {
		t.Errorf("error was not expected while getting doc: %s", err)
	}

	if res.Title != "Test Doc" {
		t.Errorf("expected title 'Test Doc', got '%s'", res.Title)
	}
	if res.Content != "Hello World" {
		t.Errorf("expected content 'Hello World', got '%s'", res.Content)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies that PostgreSQL Store updates concept content with row locking (SELECT ... FOR UPDATE) and history logging.
// [Execution Surface]: In-Memory PostgreSQL Mock Pool (pgxmock)
// [Assertions]: Transaction locks row, updates concepts, and writes new history version.
// -----------------------------------------------------------------------------
func TestStore_Update(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	store := &Store{pool: mock}

	doc := gyrus.Concept{
		ID:      "doc-1",
		Version: 1,
		Title:   "Test Doc", Tenant: "root", Name: "root/reference/doc-1",
	}
	fmBytes, _ := json.Marshal(doc)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT frontmatter, content FROM concepts WHERE id = \\$1 FOR UPDATE").
		WithArgs("root/reference/doc-1").
		WillReturnRows(pgxmock.NewRows([]string{"frontmatter", "content"}).AddRow(fmBytes, "Old content"))

	mock.ExpectExec("UPDATE concepts SET frontmatter = \\$1, content = \\$2 WHERE id = \\$3").
		WithArgs(pgxmock.AnyArg(), "New content", "root/reference/doc-1").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectExec("INSERT INTO concepts_history").
		WithArgs("root/reference/doc-1", 2, pgxmock.AnyArg(), "New content", pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mock.ExpectCommit()

	patchTitle := "Updated Test Doc"
	patchContent := "New content"
	patch := gyrus.ConceptPatch{
		Title:   &patchTitle,
		Content: &patchContent,
	}

	ref, err := store.Update(context.Background(), "root/reference/doc-1", patch, 1)
	if err != nil {
		t.Errorf("error was not expected while updating doc: %s", err)
	}

	if ref.Version != 2 {
		t.Errorf("expected version 2, got %d", ref.Version)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies that PostgreSQL Store upserts and queries graph relationship edges.
// [Execution Surface]: In-Memory PostgreSQL Mock Pool (pgxmock)
// [Assertions]: INSERT and SELECT queries on concept_edges table execute cleanly.
// -----------------------------------------------------------------------------
func TestStore_GraphEdges(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	store := &Store{pool: mock}

	// Test UpsertEdges
	edges := []gyrus.ConceptEdge{
		{
			FromConceptID:    "root/reference/doc-1",
			ToConceptID:      "root/reference/doc-2",
			RelationshipType: gyrus.RelDependsOn,
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO concept_edges").
		WithArgs("root/reference/doc-1", "root/reference/doc-2", gyrus.RelDependsOn).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	err = store.UpsertEdges(context.Background(), edges)
	if err != nil {
		t.Errorf("error was not expected while upserting edges: %s", err)
	}

	// Test Neighbors
	rows := pgxmock.NewRows([]string{"from_concept_id", "to_concept_id", "relationship_type"}).
		AddRow("root/reference/doc-1", "root/reference/doc-2", string(gyrus.RelDependsOn))

	mock.ExpectQuery("SELECT from_concept_id, to_concept_id, relationship_type FROM concept_edges WHERE from_concept_id = \\$1 OR to_concept_id = \\$1").
		WithArgs("root/reference/doc-1").
		WillReturnRows(rows)

	neighbors, err := store.Neighbors(context.Background(), "root/reference/doc-1", gyrus.EdgeFilter{})
	if err != nil {
		t.Errorf("error was not expected while getting neighbors: %s", err)
	}

	if len(neighbors) != 1 {
		t.Errorf("expected 1 neighbor, got %d", len(neighbors))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies that PostgreSQL Store implements gyrus.SchemaStore for schema CRUD operations in the schemas table.
// [Execution Surface]: In-Memory PostgreSQL Mock Pool (pgxmock)
// [Assertions]: Executes INSERT/ON CONFLICT for SaveSchema, SELECT for GetSchema, SELECT for ListSchemas, and DELETE for DeleteSchema.
// -----------------------------------------------------------------------------
func TestStore_SchemaStore(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	store := &Store{pool: mock}
	ctx := context.Background()

	content := `---
id: <unique-id>
title: <Title>
category: operations
type: runbook
tenant: ops
version: 1
status: active
---

# Runbook
`

	// 1. SaveSchema
	mock.ExpectExec("INSERT INTO schemas").
		WithArgs("runbook", content).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := store.SaveSchema(ctx, "runbook", content); err != nil {
		t.Errorf("SaveSchema failed: %v", err)
	}

	// 2. GetSchema
	schemaRows := pgxmock.NewRows([]string{"content"}).AddRow(content)
	mock.ExpectQuery("SELECT content FROM schemas WHERE id = \\$1").
		WithArgs("runbook").
		WillReturnRows(schemaRows)

	fetched, err := store.GetSchema(ctx, "runbook")
	if err != nil {
		t.Errorf("GetSchema failed: %v", err)
	}
	if fetched != content {
		t.Errorf("GetSchema content mismatch")
	}

	// 3. ListSchemas
	listRows := pgxmock.NewRows([]string{"id"}).AddRow("adr").AddRow("runbook")
	mock.ExpectQuery("SELECT id FROM schemas ORDER BY id").
		WillReturnRows(listRows)

	list, err := store.ListSchemas(ctx)
	if err != nil {
		t.Errorf("ListSchemas failed: %v", err)
	}
	if len(list) != 2 || list[0] != "adr" || list[1] != "runbook" {
		t.Errorf("ListSchemas unexpected result: %v", list)
	}

	// 4. DeleteSchema
	mock.ExpectExec("DELETE FROM schemas WHERE id = \\$1").
		WithArgs("runbook").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	if err := store.DeleteSchema(ctx, "runbook"); err != nil {
		t.Errorf("DeleteSchema failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
