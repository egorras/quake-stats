package main

import (
	"context"
	"encoding/json"
	"log"
	"runtime"
	"time"
)

// Event represents a game event
type Event struct {
	Type string          `json:"TYPE"`
	Data json.RawMessage `json:"DATA"`
}

// EventProcessor handles batching and processing of events
type EventProcessor struct {
	config     Config
	eventChan  chan Event
	buffer     []Event
	bufferSize int
	dbClient   *PostgresClient
	stats      struct {
		eventsProcessed  int64
		batchesProcessed int64
		lastReportTime   time.Time
	}
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(cfg Config, dbClient *PostgresClient) *EventProcessor {
	return &EventProcessor{
		config:     cfg,
		eventChan:  make(chan Event, 1000),
		buffer:     make([]Event, 0, cfg.BatchSize),
		bufferSize: cfg.BatchSize,
		dbClient:   dbClient,
		stats: struct {
			eventsProcessed  int64
			batchesProcessed int64
			lastReportTime   time.Time
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

// Run implements the EventProcessor interface for testing
func (p *EventProcessor) Run(ctx context.Context) {
	p.Process(ctx)
}

// ProcessEvent implements the EventProcessor interface for testing
func (p *EventProcessor) ProcessEvent(e Event) {
	p.Submit(e)
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
	
	// Store events in PostgreSQL if enabled
	if p.dbClient != nil {
		if err := p.dbClient.StoreEvents(p.buffer); err != nil {
			log.Printf("Error storing events in PostgreSQL: %v", err)
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