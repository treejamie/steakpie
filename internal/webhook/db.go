package webhook

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// EventStore manages webhook event deduplication
type EventStore struct {
	db *sql.DB
}

// NewEventStore creates a new EventStore with the given database path.
// If dbPath is empty, defaults to "db.sqlite".
// Use ":memory:" for in-memory databases (testing).
func NewEventStore(dbPath string) (*EventStore, error) {
	if dbPath == "" {
		dbPath = "db.sqlite"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Create table if not exists
	if err := initSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &EventStore{db: db}, nil
}

// initSchema creates the events table if it doesn't exist
func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS events (
		delivery_id TEXT PRIMARY KEY,
		tag TEXT NOT NULL,
		version_id INTEGER NOT NULL,
		sha TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		repository TEXT NOT NULL
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_events_content_dedup ON events(tag, version_id, sha);
	`

	_, err := db.Exec(schema)
	return err
}

// RecordEvent attempts to record a webhook event.
// Returns true if the event is new (successfully inserted).
// Returns false if the event is a duplicate (same tag+version_id+sha content).
// Returns error for other database failures.
func (es *EventStore) RecordEvent(deliveryID, tag string, versionID int64, sha, repository string) (bool, error) {
	tx, err := es.db.Begin()
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check for existing row with same content
	var exists int
	err = tx.QueryRow(
		`SELECT 1 FROM events WHERE tag = ? AND version_id = ? AND sha = ?`,
		tag, versionID, sha,
	).Scan(&exists)
	if err == nil {
		// Row found — content duplicate
		return false, nil
	}
	if err != sql.ErrNoRows {
		return false, fmt.Errorf("failed to check for duplicate: %w", err)
	}

	// Insert new event
	_, err = tx.Exec(
		`INSERT INTO events (delivery_id, tag, version_id, sha, timestamp, repository) VALUES (?, ?, ?, ?, ?, ?)`,
		deliveryID, tag, versionID, sha, time.Now().UTC(), repository,
	)
	if err != nil {
		// Constraint error as race-condition safety net
		if isConstraintError(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to insert event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return true, nil
}

// isConstraintError checks if the error is a UNIQUE constraint violation
func isConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "UNIQUE constraint failed") ||
		strings.Contains(errMsg, "constraint failed")
}

// Close closes the database connection
func (es *EventStore) Close() error {
	return es.db.Close()
}

// Stats returns statistics about stored events (useful for monitoring)
func (es *EventStore) Stats() (total int, err error) {
	err = es.db.QueryRow("SELECT COUNT(*) FROM events").Scan(&total)
	return
}
