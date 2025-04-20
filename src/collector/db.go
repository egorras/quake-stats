package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
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
	db              *sql.DB
	tableName       string
	config          Config
	lastActivity    time.Time
	connectionMutex sync.Mutex
	closed          bool
	idleTimeout     time.Duration
	checkTimer      *time.Timer
}

// NewPostgresClient creates a new PostgreSQL client
func NewPostgresClient(cfg Config) (DBClient, error) {
	if !cfg.PostgresEnabled {
		return nil, nil
	}

	client := &PostgresClient{
		tableName:    cfg.PostgresTable,
		config:       cfg,
		lastActivity: time.Now(),
		idleTimeout:  5 * time.Minute, // Default idle timeout - 5 minutes
	}

	// Connect immediately for the first time
	if err := client.connect(); err != nil {
		return nil, err
	}

	// Start the connection checker
	client.startConnectionChecker()

	log.Printf("Successfully connected to PostgreSQL database")
	return client, nil
}

// connect establishes a connection to the database
func (p *PostgresClient) connect() error {
	p.connectionMutex.Lock()
	defer p.connectionMutex.Unlock()

	if p.db != nil && !p.closed {
		// Connection already exists and is not marked as closed
		return nil
	}

	// Use the connection string directly
	connectionString := p.config.PostgresConnectionString

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	p.db = db
	p.closed = false
	p.lastActivity = time.Now()

	// Ensure the events table exists
	if err := p.createTable(); err != nil {
		p.db.Close()
		p.db = nil
		return fmt.Errorf("failed to create table: %w", err)
	}

	log.Printf("Database connection established")
	return nil
}

// startConnectionChecker periodically checks for idle connections
func (p *PostgresClient) startConnectionChecker() {
	p.checkTimer = time.AfterFunc(30*time.Second, func() {
		p.checkAndCloseIdleConnection()
	})
}

// checkAndCloseIdleConnection checks if the connection has been idle for too long
func (p *PostgresClient) checkAndCloseIdleConnection() {
	p.connectionMutex.Lock()
	defer p.connectionMutex.Unlock()

	if p.db == nil || p.closed {
		// Connection already closed
		p.checkTimer = time.AfterFunc(30*time.Second, func() {
			p.checkAndCloseIdleConnection()
		})
		return
	}

	idleTime := time.Since(p.lastActivity)
	if idleTime > p.idleTimeout {
		log.Printf("Closing idle database connection after %v of inactivity", idleTime.Round(time.Second))
		p.db.Close()
		p.db = nil
		p.closed = true
	}

	// Schedule next check
	p.checkTimer = time.AfterFunc(30*time.Second, func() {
		p.checkAndCloseIdleConnection()
	})
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

	// Ensure we have a connection
	p.connectionMutex.Lock()
	if p.db == nil || p.closed {
		if err := p.connect(); err != nil {
			p.connectionMutex.Unlock()
			return fmt.Errorf("failed to connect to database: %w", err)
		}
	}
	// Update last activity time
	p.lastActivity = time.Now()
	p.connectionMutex.Unlock()

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
	p.connectionMutex.Lock()
	defer p.connectionMutex.Unlock()

	// Stop the timer
	if p.checkTimer != nil {
		p.checkTimer.Stop()
	}

	if p.db != nil && !p.closed {
		p.closed = true
		return p.db.Close()
	}
	return nil
} 