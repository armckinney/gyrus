package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port   int
	DBAddr string
}

func Load(filenames ...string) (*Config, error) {
	// Load .env file if it exists, but don't fail if it doesn't
	_ = godotenv.Load(filenames...)

	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 8080 // Default port
	}

	dbAddr := os.Getenv("DB_URL")
	if dbAddr == "" {
		return nil, fmt.Errorf("DB_URL environment variable is required")
	}

	return &Config{
		Port:   port,
		DBAddr: dbAddr,
	}, nil
}
