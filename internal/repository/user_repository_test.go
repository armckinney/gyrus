package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUserRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "username", "email"}).
		AddRow(1, "alice", "alice@example.com").
		AddRow(2, "bob", "bob@example.com")

	mock.ExpectQuery("^SELECT id, username, email FROM users$").WillReturnRows(rows)

	users, err := repo.GetAll(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	if users[0].Username != "alice" {
		t.Errorf("Expected first user to be alice, got %s", users[0].Username)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
