package db

import (
	"os"
	"testing"
	"strings"
	"bugtracker-backend/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseInitialization(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	bug := &models.Bug{
		Title:       "Test",
		Description: "Test",
	}

	err := CreateBug(bug)
	assert.NoError(t, err)
	assert.NotZero(t, bug.ID)
}

func TestMultipleInitializations(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	err := Init()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database already initialized")
}

func TestCleanup(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	Cleanup()

	// Test DB is inaccessible after cleanup
	bug := &models.Bug{Title: "Test", Description: "Test"}
	err := CreateBug(bug)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not initialized",
		"Should get 'database not initialized' error after cleanup")
}

func TestInitWithInvalidDSN(t *testing.T) {
	Cleanup()
	
	env := strings.ToLower(os.Getenv("APP_ENV"))
	invalidDSN := "postgres://invalid:invalid@localhost:1/invalid?sslmode=disable"

	if env == "production" {
		t.Setenv("DATABASE_URL", invalidDSN)
	} else {
		t.Setenv("TEST_DATABASE_URL", invalidDSN)
		t.Setenv("DATABASE_URL", invalidDSN)
	}

	err := Init()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to ping database")

	Cleanup()
}

func TestConcurrentInitializations(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	err := Init()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database already initialized")
}

func TestMain(m *testing.M) {
	code := m.Run()
	CleanupTestDB()
	os.Exit(code)
}
