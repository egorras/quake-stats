package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// MockEventProcessor implements a simplified event processor for testing
type MockEventProcessor struct {
	events []Event
}

func NewMockEventProcessor() *MockEventProcessor {
	return &MockEventProcessor{
		events: make([]Event, 0),
	}
}

// Submit implements the event submission method
func (m *MockEventProcessor) Submit(event Event) {
	m.events = append(m.events, event)
}

func (m *MockEventProcessor) ProcessEvent(event Event) {
	m.events = append(m.events, event)
}

func (m *MockEventProcessor) Run(ctx context.Context) {
	// Do nothing - this is a mock
}

// GetChannel returns a dummy channel for compatibility
func (m *MockEventProcessor) GetChannel() chan<- Event {
	ch := make(chan Event)
	go func() {
		for e := range ch {
			m.Submit(e)
		}
	}()
	return ch
}

func TestZmqCollectorCreate(t *testing.T) {
	// Test creating a ZMQ collector with valid endpoint
	processor := NewMockEventProcessor()
	collector, err := NewZmqCollector("tcp://localhost:5555", processor)
	
	if err != nil {
		t.Fatalf("Failed to create ZMQ collector: %v", err)
	}
	
	if collector == nil {
		t.Fatal("Collector should not be nil")
	}
}

func TestZmqCollectorInvalidEndpoint(t *testing.T) {
	// Test with invalid endpoint
	processor := NewMockEventProcessor()
	_, err := NewZmqCollector("invalid://endpoint", processor)
	
	if err == nil {
		t.Fatal("Expected error for invalid endpoint, got nil")
	}
}

func TestEventExtraction(t *testing.T) {
	// Test cases for extracting events from message data
	testCases := []struct {
		name     string
		message  []byte
		expected bool
		eventType string
	}{
		{
			name:     "Valid game event",
			message:  []byte(`{"TYPE":"kill","DATA":{"attacker":"Player1","victim":"Player2"}}`),
			expected: true,
			eventType: "kill",
		},
		{
			name:     "Invalid JSON",
			message:  []byte(`{not valid json}`),
			expected: false,
			eventType: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var event Event
			err := json.Unmarshal(tc.message, &event)
			
			if tc.expected {
				if err != nil {
					t.Errorf("Expected successful unmarshalling, got error: %v", err)
				}
				
				if event.Type != tc.eventType {
					t.Errorf("Expected event type to be %s, got %s", tc.eventType, event.Type)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error when unmarshalling invalid data, got none")
				}
			}
		})
	}
}

// This test would require an actual ZeroMQ server to connect to,
// so it's skipped. In a real test environment, you might use
// a test ZMQ server or a more sophisticated mock.
func TestIntegrationWithZmq(t *testing.T) {
	t.Skip("Skipping integration test that requires a ZMQ server")

	processor := NewMockEventProcessor()
	collector, err := NewZmqCollector("tcp://localhost:5555", processor) // Test server
	if err != nil {
		t.Fatalf("Failed to create ZMQ collector: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Run collector in background
	go func() {
		collector.Run(ctx)
	}()

	// Wait for some time to potentially receive messages
	time.Sleep(3 * time.Second)

	// Check if any events were processed
	// Note: This test is dependent on external factors
	t.Logf("Received %d events", len(processor.events))
} 