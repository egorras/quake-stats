package main

import (
	"context"
	"testing"
	"time"
)

func TestEventProcessorFlushBySize(t *testing.T) {
	// Create a processor with small flush size
	config := &Config{
		Events: EventsConfig{
			FlushSize:     3,
			FlushInterval: 30, // Set high to ensure we're testing size-based flushing
		},
	}

	// Channel to track when flushes happen
	flushedEvents := make(chan []Event)

	// Create a processor with a custom flush function
	processor := NewEventProcessor(config, func(events []Event) {
		flushedEvents <- events
	})

	// Start the processor
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go processor.Run(ctx)

	// Send events
	events := []Event{
		{Type: "kill", Data: map[string]interface{}{"weapon": "shotgun"}},
		{Type: "death", Data: map[string]interface{}{"cause": "fall"}},
		{Type: "join", Data: map[string]interface{}{"player": "player1"}},
	}

	for _, e := range events {
		processor.ProcessEvent(e)
	}

	// Wait for flush
	select {
	case flushed := <-flushedEvents:
		if len(flushed) != 3 {
			t.Errorf("Expected 3 events to be flushed, got %d", len(flushed))
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for flush by size")
	}
}

func TestEventProcessorFlushByTime(t *testing.T) {
	// Create a processor with small flush interval
	config := &Config{
		Events: EventsConfig{
			FlushSize:     100, // Set high to ensure we're testing time-based flushing
			FlushInterval: 1,   // 1 second
		},
	}

	// Channel to track when flushes happen
	flushedEvents := make(chan []Event)

	// Create a processor with a custom flush function
	processor := NewEventProcessor(config, func(events []Event) {
		flushedEvents <- events
	})

	// Start the processor
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go processor.Run(ctx)

	// Send a single event
	processor.ProcessEvent(Event{
		Type: "kill",
		Data: map[string]interface{}{"weapon": "shotgun"},
	})

	// Wait for time-based flush
	select {
	case flushed := <-flushedEvents:
		if len(flushed) != 1 {
			t.Errorf("Expected 1 event to be flushed, got %d", len(flushed))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for flush by time")
	}
}

func TestEventProcessorContextCancellation(t *testing.T) {
	// Create a processor
	config := &Config{
		Events: EventsConfig{
			FlushSize:     10,
			FlushInterval: 10,
		},
	}

	processor := NewEventProcessor(config, func(events []Event) {})

	// Start the processor with a context we'll cancel
	ctx, cancel := context.WithCancel(context.Background())
	processorDone := make(chan struct{})

	go func() {
		processor.Run(ctx)
		close(processorDone)
	}()

	// Cancel the context, which should stop the processor
	cancel()

	// Wait for processor to exit, with timeout
	select {
	case <-processorDone:
		// Success - processor exited
	case <-time.After(1 * time.Second):
		t.Fatal("Processor did not exit after context cancellation")
	}
} 