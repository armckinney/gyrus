package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// 1. Test Default Port
	os.Clearenv()
	os.Setenv("DB_URL", "postgres://user:pass@localhost:5432/db")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Port)
	}
	if cfg.DBAddr != "postgres://user:pass@localhost:5432/db" {
		t.Errorf("Expected DB_URL 'postgres://user:pass@localhost:5432/db', got %s", cfg.DBAddr)
	}

	// 2. Test Custom Port
	os.Setenv("PORT", "9090")
	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", cfg.Port)
	}

	// 3. Test Missing DB_URL
	os.Clearenv()
	_, err = Load()
	if err == nil {
		t.Fatal("Expected error for missing DB_URL, got nil")
	}
}
