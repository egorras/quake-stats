package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"encoding/json"

	"github.com/spf13/viper"
	zmq4 "github.com/pebbe/zmq4"
)

// Config holds all the configuration parameters
type Config struct {
	ZmqEndpoint        string
	BatchSize          int
	FlushIntervalSec   int
	VerboseLogging     bool
}

// Event represents a game event
type Event struct {
	Type string          `json:"TYPE"`
	Data json.RawMessage `json:"DATA"`
}

// loadConfig reads configuration from file and environment variables
func loadConfig() Config {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "Path to configuration file (default: config.yaml)")
	flag.Parse()

	// Initialize viper
	v := viper.New()

	// Set default values
	v.SetDefault("zmq_endpoint", "tcp://89.168.29.137:27960")
	v.SetDefault("batch_size", 10)
	v.SetDefault("flush_interval_sec", 1)
	v.SetDefault("verbose_logging", true)

	// Configure viper to read environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Configure file-based config
	if configPath != "" {
		// Set config type based on file extension
		ext := filepath.Ext(configPath)
		if ext != "" {
			v.SetConfigType(ext[1:]) // Remove the leading dot
		}

		v.SetConfigFile(configPath)
		err := v.ReadInConfig()
		if err != nil {
			log.Printf("Warning: Using default configuration values: %v", err)
		} else {
			log.Printf("Loaded configuration from %s", configPath)
		}
	}

	// Create and return the configuration
	config := Config{
		ZmqEndpoint:      v.GetString("zmq_endpoint"),
		BatchSize:        v.GetInt("batch_size"),
		FlushIntervalSec: v.GetInt("flush_interval_sec"),
		VerboseLogging:   v.GetBool("verbose_logging"),
	}

	return config
}

// logConfig logs the current configuration
func logConfig(cfg Config) {
	log.Printf("Starting collector with configuration:")
	log.Printf("- ZMQ Endpoint: %s", cfg.ZmqEndpoint)
	log.Printf("- Batch Size: %d", cfg.BatchSize)
	log.Printf("- Flush Interval: %d seconds", cfg.FlushIntervalSec)
	log.Printf("- Verbose Logging: %v", cfg.VerboseLogging)
}

// EventProcessor handles batching and processing of events
type EventProcessor struct {
	config     Config
	eventChan  chan Event
	buffer     []Event
	bufferSize int
	stats      struct {
		eventsProcessed int64
		batchesProcessed int64
		lastReportTime time.Time
	}
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(cfg Config) *EventProcessor {
	return &EventProcessor{
		config:     cfg,
		eventChan:  make(chan Event, 1000),
		buffer:     make([]Event, 0, cfg.BatchSize),
		bufferSize: cfg.BatchSize,
		stats: struct {
			eventsProcessed int64
			batchesProcessed int64
			lastReportTime time.Time
		}{0, 0, time.Now()},
	}
}

// Process starts the event processing loop
func (p *EventProcessor) Process(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(p.config.FlushIntervalSec) * time.Second)
	heartbeatTicker := time.NewTicker(30 * time.Second) // Fixed at 30s for simplicity
	defer ticker.Stop()
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.flush()
			return
		case <-ticker.C:
			p.flush()
		case <-heartbeatTicker.C:
			p.logHeartbeat()
		case e := <-p.eventChan:
			p.buffer = append(p.buffer, e)
			if len(p.buffer) >= p.bufferSize {
				p.flush()
			}
		}
	}
}

// Submit adds an event to the processing queue
func (p *EventProcessor) Submit(e Event) {
	p.eventChan <- e
}

// GetChannel returns the event channel for submitting events
func (p *EventProcessor) GetChannel() chan<- Event {
	return p.eventChan
}

// flush processes all events in the buffer
func (p *EventProcessor) flush() {
	if len(p.buffer) == 0 {
		return
	}
	
	log.Printf("Logging batch of %d events", len(p.buffer))
	if p.config.VerboseLogging {
		for i, e := range p.buffer {
			log.Printf("Event %d: Type=%s, Data=%s", i, e.Type, string(e.Data))
		}
	}
	
	// Update stats
	p.stats.eventsProcessed += int64(len(p.buffer))
	p.stats.batchesProcessed++
	
	// Clear buffer (reuse the underlying array)
	p.buffer = p.buffer[:0]
}

// logHeartbeat logs periodic stats about the collector
func (p *EventProcessor) logHeartbeat() {
	now := time.Now()
	elapsed := now.Sub(p.stats.lastReportTime).Seconds()
	
	// Calculate events per second
	eventsPerSecond := float64(p.stats.eventsProcessed) / elapsed
	batchesPerMinute := float64(p.stats.batchesProcessed) / elapsed * 60
	
	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	log.Printf("Heartbeat: Processed %.2f events/sec (%.2f batches/min), Memory: %d MB, Goroutines: %d",
		eventsPerSecond, batchesPerMinute, m.Alloc/1024/1024, runtime.NumGoroutine())
	
	// Reset stats
	p.stats.eventsProcessed = 0
	p.stats.batchesProcessed = 0
	p.stats.lastReportTime = now
}

// ZmqCollector represents a ZMQ event collector
type ZmqCollector struct {
	endpoint   string
	socket     *zmq4.Socket
	processor  *EventProcessor
	isRemote   bool
}

// NewZmqCollector creates a new ZMQ collector
func NewZmqCollector(endpoint string, processor *EventProcessor) (*ZmqCollector, error) {
	// Check if we're connecting to a remote server or binding locally
	isRemoteConnection := !strings.HasPrefix(endpoint, "tcp://127.0.0.1") && 
	                      !strings.HasPrefix(endpoint, "tcp://localhost") && 
						  !strings.HasPrefix(endpoint, "tcp://0.0.0.0")

	log.Printf("Endpoint %s is remote: %v", endpoint, isRemoteConnection)
	
	socketType := zmq4.SUB
	
	log.Printf("Creating ZMQ SUB socket for endpoint: %s", endpoint)
	socket, err := zmq4.NewSocket(socketType)
	if err != nil {
		return nil, fmt.Errorf("failed to create ZMQ socket: %w", err)
	}

	// For SUB sockets, we need to set a subscription (empty string = all messages)
	if err := socket.SetSubscribe(""); err != nil {
		socket.Close()
		return nil, fmt.Errorf("failed to set subscription: %w", err)
	}

	// Set socket options for better reliability
	if err := socket.SetLinger(0); err != nil {
		socket.Close()
		return nil, fmt.Errorf("failed to set linger: %w", err)
	}

	// Only need a short timeout since we expect regular messages
	if err := socket.SetRcvtimeo(500 * time.Millisecond); err != nil {
		socket.Close()
		return nil, fmt.Errorf("failed to set receive timeout: %w", err)
	}

	// Set socket reconnect options
	if err := socket.SetReconnectIvl(1 * time.Second); err != nil {
		socket.Close()
		return nil, fmt.Errorf("failed to set reconnect interval: %w", err)
	}
	
	if err := socket.SetReconnectIvlMax(10 * time.Second); err != nil {
		socket.Close()
		return nil, fmt.Errorf("failed to set max reconnect interval: %w", err)
	}

	return &ZmqCollector{
		endpoint:   endpoint,
		socket:     socket,
		processor:  processor,
		isRemote:   isRemoteConnection,
	}, nil
}

// Start begins collecting events from ZMQ
func (c *ZmqCollector) Start(ctx context.Context) error {
	log.Printf("Connecting to ZMQ endpoint: %s", c.endpoint)
	
	err := c.socket.Connect(c.endpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to ZMQ endpoint '%s': %w", c.endpoint, err)
	}
	
	log.Printf("Successfully connected to endpoint: %s", c.endpoint)
	log.Printf("Waiting to receive messages...")

	// Setup cleanup on context done
	go func() {
		<-ctx.Done()
		c.Close()
	}()
	
	// Main message processing loop
	receivedCount := 0
	lastLog := time.Now()
	
	for ctx.Err() == nil {
		// Periodically log that we're still waiting for messages
		if time.Since(lastLog) > 30*time.Second {
			log.Printf("Still waiting for messages (received %d so far)...", receivedCount)
			lastLog = time.Now()
		}
		
		msg, err := c.socket.RecvBytes(0)
		if err != nil {
			if err == zmq4.ETIMEDOUT {
				// This is just a timeout, continue trying
				continue
			}
			
			// Handle "resource temporarily unavailable" errors silently
			// This happens when there are no messages to receive, which is normal
			if strings.Contains(err.Error(), "resource temporarily unavailable") {
				// Just wait a bit and try again - no need to log this common condition
				time.Sleep(500 * time.Millisecond)
				continue
			}
			
			// Log other errors
			log.Printf("ZMQ receive error: %v", err)
			time.Sleep(500 * time.Millisecond) // Short delay to avoid log spam
			continue
		}

		receivedCount++

		log.Printf("Received message: %s", string(msg))
		
		if len(msg) == 0 {
			log.Printf("Received empty message, skipping")
			continue
		}
		
		var e Event
		if err := json.Unmarshal(msg, &e); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			log.Printf("Raw message: %s", string(msg))
			continue
		}

		c.processor.Submit(e)
	}

	return nil
}

// Close closes the ZMQ socket
func (c *ZmqCollector) Close() {
	if c.socket != nil {
		c.socket.Close()
		c.socket = nil
	}
}

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
	
	// Optimize for lower memory usage
	runtime.GOMAXPROCS(2) // Limit to 2 threads for free tier VM
	
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
