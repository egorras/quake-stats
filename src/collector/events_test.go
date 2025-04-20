package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestEventProcessorFlushBySize(t *testing.T) {
	// Create a processor with small flush size
	config := Config{
		BatchSize:        3, // Set small to ensure we're testing size-based flushing
		FlushIntervalSec: 30, // Set high to ensure we're testing size-based flushing
	}

	// Channel to track when flushes happen
	flushedEvents := make(chan []Event)

	// Create a mock DB client for testing
	mockClient := &mockDBClient{
		storeEventsFunc: func(events []Event) error {
			flushedEvents <- events
			return nil
		},
	}

	// Create a processor
	processor := NewEventProcessor(config, mockClient)

	// Start the processor
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go processor.Process(ctx)

	// Send events
	events := []Event{
		{Type: "kill", Data: json.RawMessage(`{"weapon": "shotgun"}`)},
		{Type: "death", Data: json.RawMessage(`{"cause": "fall"}`)},
		{Type: "join", Data: json.RawMessage(`{"player": "player1"}`)},
	}

	for _, e := range events {
		processor.Submit(e)
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
	config := Config{
		BatchSize:        100, // Set high to ensure we're testing time-based flushing
		FlushIntervalSec: 1,   // 1 second
	}

	// Channel to track when flushes happen
	flushedEvents := make(chan []Event)

	// Create a mock DB client for testing
	mockClient := &mockDBClient{
		storeEventsFunc: func(events []Event) error {
			flushedEvents <- events
			return nil
		},
	}

	// Create a processor
	processor := NewEventProcessor(config, mockClient)

	// Start the processor
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go processor.Process(ctx)

	// Send a single event
	processor.Submit(Event{
		Type: "kill",
		Data: json.RawMessage(`{"weapon": "shotgun"}`),
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
	config := Config{
		BatchSize:        10,
		FlushIntervalSec: 10,
	}

	// Create a mock DB client that does nothing
	mockClient := &mockDBClient{
		storeEventsFunc: func(events []Event) error {
			return nil
		},
	}

	processor := NewEventProcessor(config, mockClient)

	// Start the processor with a context we'll cancel
	ctx, cancel := context.WithCancel(context.Background())
	processorDone := make(chan struct{})

	go func() {
		processor.Process(ctx)
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

// mockDBClient for testing
type mockDBClient struct {
	storeEventsFunc func(events []Event) error
}

func (m *mockDBClient) StoreEvents(events []Event) error {
	if m.storeEventsFunc != nil {
		return m.storeEventsFunc(events)
	}
	return nil
}

func (m *mockDBClient) Close() error {
	return nil
}

func (m *mockDBClient) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"mock_client": true,
	}
} 