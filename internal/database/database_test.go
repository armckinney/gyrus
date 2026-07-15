package database

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHealth(t *testing.T) {
	// 1. Create a mock DB
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// 2. Create the service instance with the mock DB
	s := &service{db: db}

	// Case 1: Healthy
	mock.ExpectPing()
	// Mock stats if needed, but sqlmock doesn't mock sql.DB.Stats() return values directly
	// because Stats() calls internal runtime stats which we can't easily set via sqlmock.
	// However, the `Health` function calls `s.db.Stats()`.
	// Since `sqlmock` creates a real `*sql.DB` around a mocked driver, `Stats()` returns zero values initially.
	// We just verify that it doesn't crash and returns expected status.

	stats := s.Health()
	if stats["status"] != "up" {
		t.Errorf("expected status 'up', got '%s'", stats["status"])
	}
	if stats["message"] != "It's healthy" {
		t.Errorf("expected message 'It's healthy', got '%s'", stats["message"])
	}

	// Case 2: Unhealthy (Ping fails)
	mock.ExpectPing().WillReturnError(errors.New("db down"))

	statsv2 := s.Health()
	if statsv2["status"] != "down" {
		t.Errorf("expected status 'down', got '%s'", statsv2["status"])
	}

	// Verify that expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestClose(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	// Note: We don't defer db.Close() here because we want to test s.Close() calling it.

	s := &service{db: db}

	mock.ExpectClose()

	err = s.Close()
	if err != nil {
		t.Errorf("error was not expected while closing connection: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// Teststats logic requires manipulating the sql.DB state which is hard with just sqlmock
// because OpenConnections etc are internal counters.
// We generally assume stdlib works and test our logic around the values.
// To strictly test the threshold logic (e.g. > 40 connections), we would need
// to wrap the `Stats()` call in an interface or validly simulate load,
// which is complex for a unit test.
// For now, the basic Health check covers the main connectivity path.
