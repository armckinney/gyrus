package server

import (
	"fmt"
	"net/http"
	"time"

	"example.com/template-go/internal/database"
	"example.com/template-go/internal/handlers"
	"example.com/template-go/internal/middleware"
	"example.com/template-go/internal/repository"
)

type Server struct {
	port     int
	db       database.Service
	userRepo repository.UserRepository
}

func NewServer(port int, db database.Service, userRepo repository.UserRepository) *http.Server {
	NewServer := &Server{
		port:     port,
		db:       db,
		userRepo: userRepo,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.HelloWorld)
	mux.HandleFunc("/health", handlers.Health(s.db))
	mux.HandleFunc("/concurrency", handlers.Concurrency)

	userHandler := &handlers.UserHandler{Repo: s.userRepo}
	// Use method matching for Go 1.22+
	mux.HandleFunc("GET /users", userHandler.GetAll)
	mux.HandleFunc("POST /users", userHandler.Create)

	return middleware.Logger(mux)
}
