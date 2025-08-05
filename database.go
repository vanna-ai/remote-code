package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"web-terminal/db"

	_ "modernc.org/sqlite"
)

func initDatabase() (*sql.DB, *db.Queries) {
	dbPath := "web-terminal.db"
	
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
	return database, queries
}

func applyMigrations(database *sql.DB) error {
	// Read and execute the migration file
	migrationSQL, err := os.ReadFile("db/migrations/001_initial.sql")
	if err != nil {
		return err
	}

	_, err = database.Exec(string(migrationSQL))
	return err
}