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

func CleanupTestDB() error {
	if DB != nil {
		query := `TRUNCATE TABLE comments, bugs RESTART IDENTITY CASCADE`
		if _, err := DB.Exec(query); err != nil {
			return fmt.Errorf("failed to clean up test database: %w", err)
		}
	}

	return nil
}

/*

package db

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"bugtracker-backend/internal/models"

	"go.etcd.io/bbolt"
"github.com/lib/pq"
)

var (
	db             *bbolt.DB
	initialized    bool
	bugsBucket     = []byte("bugs")
	commentsBucket = []byte("comments")
	counterBucket  = []byte("counter")
	databasePath   = getDBPath()
)

func getDBPath() string {
	if path := os.Getenv("DB_PATH"); path != "" {
		return path
	}
	return "bugs.db"
}

func Init() error {
	if initialized {
		return fmt.Errorf("database already initialized")
	}

	var err error
	db, err = bbolt.Open(databasePath, 0600, nil)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bugsBucket)
		if err != nil {
			return fmt.Errorf("create bugs bucket: %w", err)
		}

		_, err = tx.CreateBucketIfNotExists(commentsBucket)
		if err != nil {
			return fmt.Errorf("create comments bucket: %w", err)
		}

		b, err := tx.CreateBucketIfNotExists(counterBucket)
		if err != nil {
			return fmt.Errorf("create counter bucket: %w", err)
		}
		if b.Get([]byte("bug_id")) == nil {
			if err := b.Put([]byte("bug_id"), itob(0)); err != nil {
				return fmt.Errorf("initialize bug counter: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create buckets: %w", err)
	}

	log.Println("Database initialized successfully.")
	initialized = true
	return nil
}

func CreateBug(bug *models.Bug) error {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bugsBucket)

		nextID, err := getNextID(tx)
		if err != nil {
			return err
		}

		bug.ID = nextID

		encoded, err := json.Marshal(bug)
		if err != nil {
			return fmt.Errorf("failed to marshal bug: %w", err)
		}

		return b.Put(itob(bug.ID), encoded)
	})
}

func GetBug(id int) (*models.Bug, error) {
	var bug models.Bug

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bugsBucket)
		data := b.Get(itob(id))
		if data == nil {
			return fmt.Errorf("bug not found")
		}

		return json.Unmarshal(data, &bug)
	})

	if err != nil {
		return nil, err
	}

	return &bug, nil
}

func GetAllBugs() ([]*models.Bug, error) {
	var bugs []*models.Bug

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bugsBucket)

		return b.ForEach(func(k, v []byte) error {
			var bug models.Bug
			if err := json.Unmarshal(v, &bug); err != nil {
				return fmt.Errorf("failed to unmarshal bug %s: %w", k, err)
			}
			bugs = append(bugs, &bug)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return bugs, nil
}

func DeleteBug(id int) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bugsBucket)
		if b == nil {
			return fmt.Errorf("bucket not found")
		}

		if b.Get(itob(id)) == nil {
			return fmt.Errorf("bug not found")
		}

		return b.Delete(itob(id))
	})
}

func Cleanup() {
	if db != nil {
		db.Close()
		db = nil
	}
	initialized = false
}

func getNextID(tx *bbolt.Tx) (int, error) {
	b := tx.Bucket(counterBucket)
	id := b.Get([]byte("lastBugID"))

	var nextID int
	if id == nil {
		nextID = 1
	} else {
		nextID = btoi(id) + 1
	}

	if err := b.Put([]byte("lastBugID"), itob(nextID)); err != nil {
		return 0, err
	}

	return nextID, nil
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func btoi(b []byte) int {
	return int(binary.BigEndian.Uint64(b))
}

func UpdateBug(bug *models.Bug) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bugsBucket)

		existing := b.Get(itob(bug.ID))
		if existing == nil {
			return fmt.Errorf("bug not found")
		}

		bug.UpdatedAt = time.Now()

		encoded, err := json.Marshal(bug)
		if err != nil {
			return fmt.Errorf("failed to marshal bug: %w", err)
		}

		return b.Put(itob(bug.ID), encoded)
	})
}

func CleanupTestDB() error {
	if db == nil {
		return nil
	}

	err := db.Update(func(tx *bbolt.Tx) error {
		if err := tx.DeleteBucket(bugsBucket); err != nil {
			return err
		}
		if err := tx.DeleteBucket(commentsBucket); err != nil {
			return err
		}
		if err := tx.DeleteBucket(counterBucket); err != nil {
			return err
		}

		if _, err := tx.CreateBucket(bugsBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucket(commentsBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucket(counterBucket); err != nil {
			return err
		}
		return nil
	})

	return err
}

func DeleteAllBugs() (int, error) {
	var count int
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bugsBucket)
		if b == nil {
			count = 0
			return nil
		}
		count = b.Stats().KeyN

		if err := tx.DeleteBucket(bugsBucket); err != nil {
			return fmt.Errorf("delete bugs bucket: %w", err)
		}
		
		if _, err := tx.CreateBucket(bugsBucket); err != nil {
			return fmt.Errorf("create bugs bucket: %w", err)
		}

		c := tx.Bucket(counterBucket)
		if err := c.Put([]byte("lastBugID"), itob(0)); err != nil {
			return fmt.Errorf("reset bug counter: %w", err)
		}

		return nil
	})
	return count, err
}

*/