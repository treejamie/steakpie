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
		id TEXT PRIMARY KEY,
		timestamp DATETIME NOT NULL,
		repository TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_events_repository ON events(repository);
	`

	_, err := db.Exec(schema)
	return err
}

// RecordEvent attempts to record a webhook event.
// Returns true if the event is new (successfully inserted).
// Returns false if the event is a duplicate (constraint violation).
// Returns error for other database failures.
func (es *EventStore) RecordEvent(deliveryID, repository string) (bool, error) {
	query := `INSERT INTO events (id, timestamp, repository) VALUES (?, ?, ?)`

	_, err := es.db.Exec(query, deliveryID, time.Now().UTC(), repository)
	if err != nil {
		// Check if this is a constraint violation (duplicate)
		if isConstraintError(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to insert event: %w", err)
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
