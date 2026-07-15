# Go API Template

This is a template for a Go API project. It follows a standard directory structure and includes basic setup for an HTTP server, database connection (PostgreSQL), and routing.

## Project Structure

- `cmd/api`: Entry point of the application.
- `internal/server`: Server setup and route registration.
- `internal/database`: Database connection logic.

## Prerequisites

- Go 1.25+
- Docker (for database)

## Getting Started

1. **Clone the repository** (if not already done).


2. **Environment Variables**:
    The application looks for a `.env` file (loaded via `godotenv`). You can create one based on your needs.
    Example `.env`:

    ```env
    PORT=8080
    DB_URL=postgres://postgres:postgres@localhost:5432/templatedb?sslmode=disable
    ```

    *Note: The project uses `github.com/joho/godotenv` to auto-load `.env` files locally.*

3. **Run the Database**:
    If you are running within the provided Devcontainer, a PostgreSQL instance is already running and configured (service `devdatabase`).
    
    If running locally without the Devcontainer, ensure you have a PostgreSQL instance running and update `DB_URL` accordingly.

4. **Run the Application**:

    ```bash
    make run
    ```

    The server will start on `http://localhost:8080` (or the port specified in `.env`).

5. **Test the Endpoints**:
    - Health Check: `http://localhost:8080/health`
    - Hello World: `http://localhost:8080/`

## Useful Commands

- `make build`: Build the binary.
- `make run`: Run the application.
- `make test`: Run tests (includes database setup).
- `make lint`: Run linters (`gofmt` and `go vet`).
- `make fmt`: Format code (`go fmt`).
- `make watch`: Run with live reload (uses `air`).
- `make clean`: Clean up artifacts.

## Dependencies

- [chi](https://github.com/go-chi/chi) or Standard `net/http` for routing (Used Standard library here).
- [pgx](https://github.com/jackc/pgx) for PostgreSQL driver.
- [godotenv](https://github.com/joho/godotenv) for environment variable management.

# template-go
