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
	GetMetrics() map[string]interface{}
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
	// Connection metrics
	connectCount        int
	disconnectCount     int
	reconnectCount      int
	totalEventsStored   int
	lastConnectTime     time.Time
	lastDisconnectTime  time.Time
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
		idleTimeout:  time.Duration(cfg.PostgresIdleTimeoutMin) * time.Minute,
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

	// Track reconnect count if this isn't the first connection
	if p.connectCount > 0 {
		p.reconnectCount++
	}

	// Use the connection string directly
	connectionString := p.config.PostgresConnectionString

	// Try to connect with exponential backoff
	var db *sql.DB
	var err error
	maxRetries := 3
	retryDelay := 500 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", connectionString)
		if err != nil {
			log.Printf("Failed to open database connection (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
			continue
		}

		// Test the connection
		if err := db.Ping(); err != nil {
			log.Printf("Failed to ping database (attempt %d/%d): %v", i+1, maxRetries, err)
			db.Close()
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
			continue
		}

		// Connection successful
		break
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	p.db = db
	p.closed = false
	p.lastActivity = time.Now()
	p.connectCount++
	p.lastConnectTime = time.Now()

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
		p.disconnectCount++
		p.lastDisconnectTime = time.Now()
	}

	// Schedule next check
	p.checkTimer = time.AfterFunc(30*time.Second, func() {
		p.checkAndCloseIdleConnection()
	})
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
		"INSERT INTO %s (EventType, EventData) VALUES ($1, $2)",
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

	// Update metrics
	p.totalEventsStored += len(events)

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

// GetMetrics returns the connection metrics
func (p *PostgresClient) GetMetrics() map[string]interface{} {
	p.connectionMutex.Lock()
	defer p.connectionMutex.Unlock()

	metrics := map[string]interface{}{
		"connect_count":        p.connectCount,
		"disconnect_count":     p.disconnectCount,
		"reconnect_count":      p.reconnectCount,
		"total_events_stored":  p.totalEventsStored,
		"connection_active":    p.db != nil && !p.closed,
		"idle_timeout_minutes": p.idleTimeout.Minutes(),
	}

	if !p.lastConnectTime.IsZero() {
		metrics["last_connect_time"] = p.lastConnectTime.Format(time.RFC3339)
	}
	if !p.lastDisconnectTime.IsZero() {
		metrics["last_disconnect_time"] = p.lastDisconnectTime.Format(time.RFC3339)
	}

	return metrics
} 