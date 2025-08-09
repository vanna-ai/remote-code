package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"remote-code/db"

	_ "modernc.org/sqlite"
)

func initDatabase() (*sql.DB, *db.Queries) {
	database, queries, _ := initDatabaseWithPathAndReturn("remote-code.db")
	return database, queries
}

func initTestDatabase() (*sql.DB, *db.Queries, string) {
	// Use a unique filename for each test run to avoid conflicts
	testDbPath := fmt.Sprintf("remote-code-test-%d.db", time.Now().UnixNano())
	return initDatabaseWithPathAndReturn(testDbPath)
}

func initDatabaseWithPathAndReturn(dbPath string) (*sql.DB, *db.Queries, string) {
	
	// Ensure directory exists
	dbDir := filepath.Dir(dbPath)
	if dbDir != "." {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			log.Fatalf("Failed to create database directory: %v", err)
		}
	}

	// Open database connection
	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Test the connection
	if err := database.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Apply migrations
	if err := applyMigrations(database); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	queries := db.New(database)
	return database, queries, dbPath
}

func applyMigrations(database *sql.DB) error {
	migrations := []string{
		"db/migrations/001_initial.sql",
		"db/migrations/002_elo_tracking.sql",
	}

	for _, migrationPath := range migrations {
		migrationSQL, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %v", migrationPath, err)
		}

		_, err = database.Exec(string(migrationSQL))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %v", migrationPath, err)
		}
	}

	return nil
}