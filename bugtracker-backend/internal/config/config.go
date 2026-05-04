package config

import (
	"os"
	"path/filepath"
	"fmt"
	"github.com/joho/godotenv"
	"strings"
)

func GetDatabaseURL() string {
	env := strings.ToLower(os.Getenv("APP_ENV"))

	switch env {
		case "production":
			dsn := os.Getenv("DATABASE_URL")
			if dsn == "" {
				panic("DATABASE_URL must be set in production")
			}
			return dsn

		default: // local, staging
			testDSN := os.Getenv("TEST_DATABASE_URL")
			if testDSN != "" {
				return testDSN
			}

			panic("TEST_DATABASE_URL must be set for local & staging environments")
	}
}

func LoadEnvFromRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := wd
	for {
		envPath := filepath.Join(dir, ".env")

		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Load(envPath); err != nil {
				return "", err
			}
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("No .env file found, using environment variables")
}