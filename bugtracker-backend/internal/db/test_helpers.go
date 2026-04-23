package db

import (
	"os"
	"testing"
	"path/filepath"
	"github.com/joho/godotenv"
)

func SetupTestDB(t *testing.T) func() {
	t.Helper()
	
	// Load environment variables from .env file
	if os.Getenv("DATABASE_URL") == "" && os.Getenv("TEST_DATABASE_URL") == "" {
		loadEnvFromRoot(t)
	}

	testDSN := os.Getenv("TEST_DATABASE_URL")
	if testDSN == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	originalDSN := os.Getenv("DATABASE_URL")
	_ = os.Setenv("DATABASE_URL", testDSN)

	if initialized {
		Cleanup()
	}

	if err := Init(); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	return func() {
		Cleanup()

		_ = os.Setenv("DATABASE_URL", originalDSN)
	}
}

func loadEnvFromRoot(t *testing.T) {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Walk up until bugtracker-backend
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, ".env")); err == nil {
			// Found .env
			if err := godotenv.Load(filepath.Join(dir, ".env")); err != nil {
				t.Fatalf("failed to load .env: %v", err)
			}
			t.Logf("Loaded .env from: %s", dir)
			return
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	t.Fatal(".env file not found in any parent directory")
}