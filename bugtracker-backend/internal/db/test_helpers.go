package db

import (
	"database/sql"
	"os"
	"testing"
	"fmt"
	"strings"
	"bugtracker-backend/internal/config"
)

var TestDB *sql.DB

func SetupTestDB(t *testing.T) func() {
	t.Helper()
	
	// Load environment variables from .env file
	if os.Getenv("DATABASE_URL") == "" && os.Getenv("TEST_DATABASE_URL") == "" {
		loadEnvFromRoot(t)
	}

	if initialized {
		Cleanup()
	}

	if err := Init(); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	return func() {
		Cleanup()
	}
}

func CleanupTestDB() error {
	if strings.ToLower(os.Getenv("APP_ENV")) == "production" {
		return fmt.Errorf("refusing to cleanup database in production")
	}

	if DB == nil {
		return nil
	}

	query := `
		TRUNCATE TABLE comments, bugs RESTART IDENTITY CASCADE;
	`

	if _, err := DB.Exec(query); err != nil {
		return fmt.Errorf("failed to clean up test database: %w", err)
	}

	return nil
}

func loadEnvFromRoot(t *testing.T) {
	t.Helper()

	dir, err := config.LoadEnvFromRoot()
	if err != nil {
		t.Fatalf("failed to load .env: %v", err)
	}

	t.Logf("Loaded .env from: %s", dir)
}