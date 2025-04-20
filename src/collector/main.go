package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

// setupSignalHandling configures graceful shutdown on SIGINT/SIGTERM
func setupSignalHandling(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("Shutting down...")
		cancel()
	}()
}

func main() {
	// Setup logging
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("Quake Stats Collector starting")
	
	// Check if we're running the import tool
	ImportTool()
	
	runtime.GOMAXPROCS(2)
	
	// Load configuration
	cfg := loadConfig()
	logConfig(cfg)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	setupSignalHandling(cancel)

	// Initialize storage clients
	var dbClients []DBClient

	// Initialize PostgreSQL client if enabled
	if cfg.PostgresEnabled {
		dbClient, err := NewPostgresClient(cfg)
		if err != nil {
			log.Printf("Warning: Failed to initialize PostgreSQL client: %v", err)
		} else if dbClient != nil {
			dbClients = append(dbClients, dbClient)
			defer dbClient.Close()
		}
	}

	// Initialize file backup client if enabled
	if cfg.FileBackupEnabled {
		fileBackupConfig := FileBackupConfig{
			Enabled:     cfg.FileBackupEnabled,
			BasePath:    cfg.FileBackupPath,
			MaxFileSize: int64(cfg.FileBackupMaxSizeMB) * 1024 * 1024, // Convert MB to bytes
			MaxFileAge:  time.Duration(cfg.FileBackupMaxAgeHours) * time.Hour,
		}
		
		fileClient, err := NewFileBackupClient(fileBackupConfig)
		if err != nil {
			log.Printf("Warning: Failed to initialize file backup client: %v", err)
		} else if fileClient != nil {
			dbClients = append(dbClients, fileClient)
			defer fileClient.Close()
		}
	}

	// Create a multi-client if we have multiple storage options
	var storageClient DBClient
	if len(dbClients) > 1 {
		storageClient = NewMultiDBClient(dbClients)
	} else if len(dbClients) == 1 {
		storageClient = dbClients[0]
	}

	// Create event processor
	processor := NewEventProcessor(cfg, storageClient)
	
	// Create ZMQ collector factory
	createZmqCollector := func(config *Config, proc EventProcessorInterface) (Collector, error) {
		return NewZmqCollector(config.ZmqEndpoint, proc)
	}

	// Create collector manager
	manager, err := NewCollectorManager(&cfg, processor, createZmqCollector)
	if err != nil {
		log.Fatalf("Failed to create collector manager: %v", err)
	}
	
	// Start the manager (this will block until context is cancelled)
	manager.Run(ctx)

	log.Println("Collector shut down")
} 