package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"bugtracker-backend/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	DB             *sql.DB
	initialized    bool
)

func Init() error {
	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	if initialized {
		return fmt.Errorf("database already initialized")
	}

	var err error
	DB, err = sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println(":green_circle: connected to Postgres (Neon)")

	if err := createTables(); err != nil {
		_ = DB.Close()
		DB = nil
		return err
	}

	initialized = true
	return nil
}

func createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS bugs (
		id SERIAL PRIMARY KEY,
		title TEXT,
		description TEXT,
		status TEXT,
		priority TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS comments (
		id SERIAL PRIMARY KEY,
		bug_id INTEGER NOT NULL REFERENCES bugs(id) ON DELETE CASCADE,
		content TEXT,
		author TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := DB.Exec(query)
	return err
}

func CreateBug(bug *models.Bug) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `
	INSERT INTO bugs (title, description, status, priority)
	VALUES ($1, $2, $3, $4)
	RETURNING id;
	`

	return DB.QueryRow(
		query,
		bug.Title,
		bug.Description,
		bug.Status,
		bug.Priority,
	).Scan(&bug.ID)
}

func GetBug(id int) (*models.Bug, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var bug models.Bug

	query := `
	SELECT id, title, description, status, priority, created_at, updated_at
	FROM bugs
	WHERE id = $1;
	`

	err := DB.QueryRow(query, id).Scan(
		&bug.ID,
		&bug.Title,
		&bug.Description,
		&bug.Status,
		&bug.Priority,
		&bug.CreatedAt,
		&bug.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("bug not found")
	}

	if err != nil {
		return nil, err
	}

	return &bug, nil
}

func GetAllBugs() ([]*models.Bug, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := DB.Query(`
		SELECT id, title, description, status, priority, created_at, updated_at
		FROM bugs
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bugs []*models.Bug

	for rows.Next() {
		var bug models.Bug
		err := rows.Scan(
			&bug.ID,
			&bug.Title,
			&bug.Description,
			&bug.Status,
			&bug.Priority,
			&bug.CreatedAt,
			&bug.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bugs = append(bugs, &bug)
	}

	return bugs, rows.Err()
}

func UpdateBug(bug *models.Bug) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `
	UPDATE bugs
	SET title = $1,
	    description = $2,
	    status = $3,
	    priority = $4,
	    updated_at = CURRENT_TIMESTAMP
	WHERE id = $5;
	`

	result, err := DB.Exec(
		query,
		bug.Title,
		bug.Description,
		bug.Status,
		bug.Priority,
		bug.ID,
	)

	rows, err := result.RowsAffected()

	if rows == 0  {
		return fmt.Errorf("bug not found")
	}

	return err
}

func DeleteBug(id int) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	result, err := DB.Exec("DELETE FROM bugs WHERE id = $1;", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("bug not found")
	}

	return err
}

func DeleteAllBugs() (int, error) {
	if DB == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	var count int

	err := DB.QueryRow("SELECT COUNT(*) FROM bugs;").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count bugs: %w", err)
	}

	_, err = DB.Exec("TRUNCATE TABLE bugs RESTART IDENTITY CASCADE;")
	if err != nil {
		return 0, fmt.Errorf("failed to delete all bugs: %w", err)
	}

	return count, err
}

func Cleanup() {
	if DB != nil {
		_ = DB.Close()
		DB = nil
	}
	initialized = false
}