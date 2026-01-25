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
	// Create migrations tracking table if it doesn't exist
	_, err := database.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %v", err)
	}

	// Check if this is an existing database that needs migration seeding
	// by checking if the migrations table is empty but schema exists
	var migrationCount int
	err = database.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount)
	if err != nil {
		return fmt.Errorf("failed to count migrations: %v", err)
	}

	if migrationCount == 0 {
		// Check if 001_initial was already applied (projects table exists)
		var tableExists int
		err = database.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='projects'").Scan(&tableExists)
		if err != nil {
			return fmt.Errorf("failed to check for projects table: %v", err)
		}
		if tableExists > 0 {
			// Seed 001 as already applied
			database.Exec("INSERT OR IGNORE INTO schema_migrations (version) VALUES (?)", "db/migrations/001_initial.sql")
		}

		// Check if 002_elo_tracking was already applied (elo_rating column exists on agents)
		var eloColumnExists int
		err = database.QueryRow("SELECT COUNT(*) FROM pragma_table_info('agents') WHERE name='elo_rating'").Scan(&eloColumnExists)
		if err != nil {
			return fmt.Errorf("failed to check for elo_rating column: %v", err)
		}
		if eloColumnExists > 0 {
			// Seed 002 as already applied
			database.Exec("INSERT OR IGNORE INTO schema_migrations (version) VALUES (?)", "db/migrations/002_elo_tracking.sql")
		}

		// Check if 003_remove_worktrees was already applied (worktrees table doesn't exist)
		var worktreesExists int
		err = database.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='worktrees'").Scan(&worktreesExists)
		if err != nil {
			return fmt.Errorf("failed to check for worktrees table: %v", err)
		}
		if worktreesExists == 0 && tableExists > 0 {
			// Worktrees table doesn't exist but other tables do, 003 was already applied
			database.Exec("INSERT OR IGNORE INTO schema_migrations (version) VALUES (?)", "db/migrations/003_remove_worktrees.sql")
		}
	}

	migrations := []string{
		"db/migrations/001_initial.sql",
		"db/migrations/002_elo_tracking.sql",
		"db/migrations/003_remove_worktrees.sql",
	}

	for _, migrationPath := range migrations {
		// Check if migration has already been applied
		var count int
		err := database.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", migrationPath).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %v", migrationPath, err)
		}
		if count > 0 {
			// Migration already applied, skip
			continue
		}

		migrationSQL, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %v", migrationPath, err)
		}

		_, err = database.Exec(string(migrationSQL))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %v", migrationPath, err)
		}

		// Record that migration was applied
		_, err = database.Exec("INSERT INTO schema_migrations (version) VALUES (?)", migrationPath)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %v", migrationPath, err)
		}
	}

	return nil
}