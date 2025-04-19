package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// DBClient interface for database operations
type DBClient interface {
	StoreEvents(events []Event) error
	Close() error
}

// PostgresClient handles database operations
type PostgresClient struct {
	db        *sql.DB
	tableName string
	config    Config
}

// NewPostgresClient creates a new PostgreSQL client
func NewPostgresClient(cfg Config) (DBClient, error) {
	if !cfg.PostgresEnabled {
		return nil, nil
	}

	// Use the connection string directly
	connectionString := cfg.PostgresConnectionString

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Hour)

	client := &PostgresClient{
		db:        db,
		tableName: cfg.PostgresTable,
		config:    cfg,
	}

	// Ensure the events table exists
	if err := client.createTable(); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	log.Printf("Successfully connected to PostgreSQL database")
	return client, nil
}

// createTable ensures the events table exists
func (p *PostgresClient) createTable() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			event_type VARCHAR(50) NOT NULL,
			event_data JSONB NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`, p.tableName)

	_, err := p.db.Exec(query)
	return err
}

// StoreEvents stores a batch of events in the database
func (p *PostgresClient) StoreEvents(events []Event) error {
	if len(events) == 0 {
		return nil
	}

	// Begin transaction
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Will be no-op if transaction is committed

	// Prepare the insert statement
	stmt, err := tx.Prepare(fmt.Sprintf(
		"INSERT INTO %s (event_type, event_data) VALUES ($1, $2)",
		p.tableName,
	))
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Insert each event
	for _, event := range events {
		_, err := stmt.Exec(event.Type, event.Data)
		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Stored %d events in PostgreSQL database", len(events))
	return nil
}

// Close closes the database connection
func (p *PostgresClient) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
} 