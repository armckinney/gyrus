package repository

import (
	"context"
	"database/sql"
	"log/slog"

	"example.com/template-go/internal/models"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	GetAll(ctx context.Context) ([]models.User, error)
	CreateUser(ctx context.Context, user models.User) (models.User, error)
}

// postgresUserRepository implements UserRepository for PostgreSQL.
type postgresUserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new PostgreSQL user repository.
func NewUserRepository(db *sql.DB) UserRepository {
	return &postgresUserRepository{
		db: db,
	}
}

// GetAll retrieves all users from the database.
func (r *postgresUserRepository) GetAll(ctx context.Context) ([]models.User, error) {
	users := []models.User{}

	query := "SELECT id, username, email FROM users"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		slog.Error("Failed to query users", "query", query, "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email); err != nil {
			slog.Error("Failed to scan user row", "error", err)
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Rows iteration error", "error", err)
		return nil, err
	}

	slog.Info("Fetched users from repository", "count", len(users))
	return users, nil
}

// CreateUser inserts a new user into the database.
func (r *postgresUserRepository) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	query := "INSERT INTO users (username, email) VALUES ($1, $2) RETURNING id"

	err := r.db.QueryRowContext(ctx, query, user.Username, user.Email).Scan(&user.ID)
	if err != nil {
		slog.Error("Failed to create user", "username", user.Username, "error", err)
		return models.User{}, err
	}

	slog.Info("User created", "id", user.ID, "username", user.Username)
	return user, nil
}
