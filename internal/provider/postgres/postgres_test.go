package postgres

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/pashagolub/pgxmock/v4"
)

func TestStore_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	store := &Store{pool: mock}
	
	doc := gyrus.Document{
		ID:      "doc-1",
		Title:   "Test Doc",
		Content: "Hello World",
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO documents").WithArgs("doc-1", pgxmock.AnyArg(), "Hello World").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("INSERT INTO documents_history").WithArgs("doc-1", 1, pgxmock.AnyArg(), "Hello World", pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("INSERT", 1))
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

func TestStore_Get(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	store := &Store{pool: mock}
	
	doc := gyrus.Document{
		ID:      "doc-1",
		Title:   "Test Doc",
	}
	fmBytes, _ := json.Marshal(doc)

	rows := pgxmock.NewRows([]string{"frontmatter", "content"}).
		AddRow(fmBytes, "Hello World")

	mock.ExpectQuery("SELECT frontmatter, content FROM documents WHERE id = \\$1").
		WithArgs("doc-1").
		WillReturnRows(rows)

	res, err := store.Get(context.Background(), "doc-1")
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

func TestStore_Update(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	store := &Store{pool: mock}
	
	doc := gyrus.Document{
		ID:      "doc-1",
		Version: 1,
		Title:   "Test Doc",
	}
	fmBytes, _ := json.Marshal(doc)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT frontmatter, content FROM documents WHERE id = \\$1 FOR UPDATE").
		WithArgs("doc-1").
		WillReturnRows(pgxmock.NewRows([]string{"frontmatter", "content"}).AddRow(fmBytes, "Old content"))
		
	mock.ExpectExec("UPDATE documents SET frontmatter = \\$1, content = \\$2 WHERE id = \\$3").
		WithArgs(pgxmock.AnyArg(), "New content", "doc-1").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectExec("INSERT INTO documents_history").
		WithArgs("doc-1", 2, pgxmock.AnyArg(), "New content", pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
		
	mock.ExpectCommit()

	patchTitle := "Updated Test Doc"
	patchContent := "New content"
	patch := gyrus.DocumentPatch{
		Title:   &patchTitle,
		Content: &patchContent,
	}

	ref, err := store.Update(context.Background(), "doc-1", patch, 1)
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

func TestStore_GraphEdges(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	store := &Store{pool: mock}

	// Test UpsertEdges
	edges := []gyrus.DocumentEdge{
		{
			FromDocumentID:   "doc-1",
			ToDocumentID:     "doc-2",
			RelationshipType: gyrus.RelDependsOn,
			CreatedBy:        "user",
			CreatedAt:        time.Now(),
		},
	}
	
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO document_edges").
		WithArgs("doc-1", "doc-2", gyrus.RelDependsOn, "user", pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	err = store.UpsertEdges(context.Background(), edges)
	if err != nil {
		t.Errorf("error was not expected while upserting edges: %s", err)
	}

	// Test Neighbors
	now := time.Now()
	rows := pgxmock.NewRows([]string{"from_document_id", "to_document_id", "relationship_type", "created_by", "created_at"}).
		AddRow("doc-1", "doc-2", string(gyrus.RelDependsOn), "user", now)

	mock.ExpectQuery("SELECT from_document_id, to_document_id, relationship_type, created_by, created_at FROM document_edges WHERE from_document_id = \\$1 OR to_document_id = \\$1").
		WithArgs("doc-1").
		WillReturnRows(rows)
		
	neighbors, err := store.Neighbors(context.Background(), "doc-1", gyrus.EdgeFilter{})
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
