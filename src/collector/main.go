package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
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
	
	runtime.GOMAXPROCS(2)
	
	// Load configuration
	cfg := loadConfig()
	logConfig(cfg)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	setupSignalHandling(cancel)

	// Create event processor
	processor := NewEventProcessor(cfg)
	go processor.Process(ctx)

	// Create and start ZMQ collector
	collector, err := NewZmqCollector(cfg.ZmqEndpoint, processor)
	if err != nil {
		log.Fatalf("Failed to create ZMQ collector: %v", err)
	}
	
	// Start collecting (this will block until context is cancelled)
	if err := collector.Start(ctx); err != nil {
		log.Fatalf("Failed to start ZMQ collector: %v", err)
	}

	log.Println("Collector shut down")
} 