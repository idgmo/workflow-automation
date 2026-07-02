package localDatabase

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3" // Loaded silently into the global package space
)

type Store struct {
	db *sql.DB
}

// NewStore launches or connects directly to the absolute target file layout path
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database channel: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database connectivity check failed: %w", err)
	}

	s := &Store{db: db}
	if err := s.bootstrapSchema(); err != nil {
		return nil, fmt.Errorf("failed to establish table patterns: %w", err)
	}

	return s, nil
}

// Close drops the system thread pool hooks safely
func (s *Store) Close() error {
	return s.db.Close()
}

// bootstrapSchema ensures tracking tables exist right away on container boot
func (s *Store) bootstrapSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS processed_events (
		event_id TEXT PRIMARY KEY,
		client_name TEXT,
		processed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := s.db.Exec(schema)
	return err
}
