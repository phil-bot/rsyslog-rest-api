package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/phil-bot/rsyslox/internal/config"
)

// DB wraps the database connection and provides helper methods.
type DB struct {
	*sql.DB
	AvailableColumns []string
	PriorityMode     PriorityMode
}

// Connect establishes a connection to the database using the TOML-based config.
func Connect(cfg *config.Config) (*DB, error) {
	dsn, err := cfg.DSN()
	if err != nil {
		return nil, fmt.Errorf("failed to build DSN: %w", err)
	}

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✓ Database connection established")

	db := &DB{DB: sqlDB}
	if err := db.initialize(); err != nil {
		return nil, err
	}

	return db, nil
}

// initialize performs initial database setup.
func (db *DB) initialize() error {
	if err := db.createIndexes(); err != nil {
		return err
	}
	if err := db.loadColumns(); err != nil {
		return err
	}
	db.PriorityMode = db.detectPriorityMode()
	return nil
}

// loadColumns loads all column names from the SystemEvents table.
// "Severity" is added as a virtual computed column.
func (db *DB) loadColumns() error {
	rows, err := db.Query("SHOW COLUMNS FROM SystemEvents")
	if err != nil {
		return fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	db.AvailableColumns = []string{}
	for rows.Next() {
		var field, colType, null, key, def, extra sql.NullString
		if err := rows.Scan(&field, &colType, &null, &key, &def, &extra); err != nil {
			log.Printf("Warning: failed to scan column info: %v", err)
			continue
		}
		if field.Valid {
			db.AvailableColumns = append(db.AvailableColumns, field.String)
		}
	}

	// Virtual column derived from Priority MOD 8
	db.AvailableColumns = append(db.AvailableColumns, "Severity")

	log.Printf("✓ Loaded %d columns from SystemEvents (+ virtual: Severity)",
		len(db.AvailableColumns)-1)
	return nil
}

// IsValidColumn checks if a column name is valid (real or virtual).
func (db *DB) IsValidColumn(column string) bool {
	for _, col := range db.AvailableColumns {
		if col == column {
			return true
		}
	}
	return false
}

// Health checks the database connection health.
func (db *DB) Health() error {
	return db.Ping()
}
