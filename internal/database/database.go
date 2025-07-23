package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return &DB{conn: conn}, nil
}

func (db *DB) DB() *sql.DB {
	return db.conn
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Begin() (*sql.Tx, error) {
	return db.conn.Begin()
}