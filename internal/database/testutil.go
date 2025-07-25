package database

import (
	"database/sql"
	"testing"

	"github.com/diegoclair/slack-rotation-bot/migrator/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *DB {
	t.Helper()

	// Create in-memory SQLite database
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "Failed to create test database")

	// Run migrations to create tables
	err = sqlite.Migrate(sqlDB)
	require.NoError(t, err, "Failed to run migrations on test database")

	return &DB{conn: sqlDB}
}

// CleanupTestDB closes the test database
func CleanupTestDB(t *testing.T, db *DB) {
	t.Helper()
	
	err := db.Close()
	require.NoError(t, err, "Failed to close test database")
}