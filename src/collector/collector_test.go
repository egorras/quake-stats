package main

import (
	"context"
	"testing"
	"time"
)

// mockEventProcessor for testing
type mockEventProcessor struct {
	processedEvents []Event
	running         bool
}

func (m *mockEventProcessor) ProcessEvent(event Event) {
	m.processedEvents = append(m.processedEvents, event)
}

func (m *mockEventProcessor) Run(ctx context.Context) {
	m.running = true
	<-ctx.Done()
	m.running = false
}

// mockCollectorImplementation for testing
type mockCollectorImplementation struct {
	started      bool
	stopped      bool
	runError     error
	eventsSent   []Event
	stopChan     chan struct{}
	processor    EventProcessorInterface
	shouldFail   bool
	failOnCreate bool
}

func newMockCollectorImplementation(processor EventProcessorInterface, shouldFail, failOnCreate bool) *mockCollectorImplementation {
	return &mockCollectorImplementation{
		stopChan:     make(chan struct{}),
		processor:    processor,
		shouldFail:   shouldFail,
		failOnCreate: failOnCreate,
	}
}

func (m *mockCollectorImplementation) Run(ctx context.Context) error {
	if m.shouldFail {
		return m.runError
	}
	
	m.started = true
	select {
	case <-ctx.Done():
		m.Stop() // Make sure Stop is called when context is cancelled
		return ctx.Err()
	case <-m.stopChan:
		return nil
	}
}

func (m *mockCollectorImplementation) Stop() {
	if !m.stopped {
		m.stopped = true
		close(m.stopChan)
	}
}

// mockCollectorFactory for testing
type mockCollectorFactory struct {
	collectors []*mockCollectorImplementation
	shouldFail bool
	failOnCreate bool
}

func (f *mockCollectorFactory) CreateCollector(config *Config, processor EventProcessorInterface) (Collector, error) {
	mockCollector := newMockCollectorImplementation(processor, f.shouldFail, f.failOnCreate)
	if mockCollector.failOnCreate {
		return nil, &CollectorError{Message: "Failed to create collector"}
	}
	f.collectors = append(f.collectors, mockCollector)
	return mockCollector, nil
}

func TestCollectorManagerCreation(t *testing.T) {
	// Test valid creation
	config := &Config{}
	processor := &mockEventProcessor{}
	factory := &mockCollectorFactory{}
	
	manager, err := NewCollectorManager(config, processor, factory.CreateCollector)
	if err != nil {
		t.Fatalf("Failed to create collector manager: %v", err)
	}
	
	if manager == nil {
		t.Fatal("Manager should not be nil")
	}
	
	if len(factory.collectors) != 1 {
		t.Fatalf("Expected 1 collector to be created, got %d", len(factory.collectors))
	}
}

func TestCollectorManagerCreationFailure(t *testing.T) {
	// Test creation failure
	config := &Config{}
	processor := &mockEventProcessor{}
	factory := &mockCollectorFactory{failOnCreate: true}
	
	_, err := NewCollectorManager(config, processor, factory.CreateCollector)
	if err == nil {
		t.Fatal("Expected error on collector creation failure, got nil")
	}
}

func TestCollectorManagerRun(t *testing.T) {
	// Test running
	config := &Config{}
	processor := &mockEventProcessor{}
	factory := &mockCollectorFactory{}
	
	manager, _ := NewCollectorManager(config, processor, factory.CreateCollector)
	
	// Run the manager in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	done := make(chan struct{})
	go func() {
		manager.Run(ctx)
		close(done)
	}()
	
	// Give it time to start
	time.Sleep(100 * time.Millisecond)
	
	// Verify processor is running
	if !processor.running {
		t.Fatal("Processor should be running")
	}
	
	// Verify collector was started
	if !factory.collectors[0].started {
		t.Fatal("Collector should be started")
	}
	
	// Cancel the context to stop
	cancel()
	
	// Wait for the manager to exit
	select {
	case <-done:
		// Success - manager exited
		// Wait a bit to ensure the collector has time to be stopped
		time.Sleep(50 * time.Millisecond)
	case <-time.After(1 * time.Second):
		t.Fatal("Manager did not exit after context cancellation")
	}
	
	// Verify collector was stopped
	if !factory.collectors[0].stopped {
		t.Fatal("Collector should be stopped")
	}
}

func TestCollectorManagerRestart(t *testing.T) {
	// Test restarting after failure
	config := &Config{}
	processor := &mockEventProcessor{}
	factory := &mockCollectorFactory{}
	
	manager, _ := NewCollectorManager(config, processor, factory.CreateCollector)
	
	// Set up the first collector to fail after running
	factory.collectors[0].shouldFail = true
	factory.collectors[0].runError = &CollectorError{Message: "Test failure", Recoverable: true}
	
	// Run the manager in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	done := make(chan struct{})
	go func() {
		manager.Run(ctx)
		close(done)
	}()
	
	// Give it time to start and fail
	time.Sleep(100 * time.Millisecond)
	
	// Trigger failure by stopping the first collector
	factory.collectors[0].Stop()
	
	// Give it time to create a new collector
	time.Sleep(100 * time.Millisecond)
	
	// Verify a second collector was created
	if len(factory.collectors) < 2 {
		t.Fatal("Expected a second collector to be created after failure")
	}
	
	// Cancel the context to stop
	cancel()
	
	// Wait for the manager to exit
	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Manager did not exit after context cancellation")
	}
} 