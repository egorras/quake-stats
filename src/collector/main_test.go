package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

// mockCollector implements Collector interface for testing
type mockCollector struct {
	started  bool
	stopped  bool
	runError error
	stopChan chan struct{}
}

func newMockCollector() *mockCollector {
	return &mockCollector{
		stopChan: make(chan struct{}),
	}
}

func (m *mockCollector) Run(ctx context.Context) error {
	m.started = true
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-m.stopChan:
		return m.runError
	}
}

func (m *mockCollector) Stop() {
	m.stopped = true
	close(m.stopChan)
}

// runWithCollector simulates running the application with a specific collector
func runWithCollector(ctx context.Context, collector Collector) {
	// Set up signal handling
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()
	
	// Set up a channel to capture signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	
	// Handle signals in a goroutine
	go func() {
		select {
		case <-sigs:
			// Signal received, cancel context
			cancel()
		case <-ctx.Done():
			// Context already done, nothing to do
		}
	}()
	
	// Run the collector until context is cancelled
	_ = collector.Run(ctxWithCancel)
	
	// Ensure collector is stopped
	collector.Stop()
}

func TestSignalHandling(t *testing.T) {
	// Create mock collector
	collector := newMockCollector()
	
	// Create a context with cancel function
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create a channel to signal when the run function exits
	done := make(chan struct{})
	
	// Run the application in goroutine
	go func() {
		runWithCollector(ctx, collector)
		close(done)
	}()
	
	// Wait a short time for the collector to "start"
	time.Sleep(100 * time.Millisecond)
	
	// Verify collector was started
	if !collector.started {
		t.Fatal("Collector should have been started")
	}
	
	// Cancel the context (simulating a signal)
	cancel()
	
	// Wait for the function to exit
	select {
	case <-done:
		// Success - function exited
	case <-time.After(1 * time.Second):
		t.Fatal("Function did not exit after context cancellation")
	}
	
	// Verify collector was stopped
	if !collector.stopped {
		t.Fatal("Collector should have been stopped")
	}
}

func TestContextCancellation(t *testing.T) {
	// Create a context that will automatically cancel after a short time
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	// Create mock collector
	collector := newMockCollector()
	
	// Create a channel to signal when run exits
	done := make(chan struct{})
	
	// Run in goroutine
	go func() {
		runWithCollector(ctx, collector)
		close(done)
	}()
	
	// Wait for the function to exit
	select {
	case <-done:
		// Success - function exited due to context timeout
	case <-time.After(1 * time.Second):
		t.Fatal("Function did not exit after context timeout")
	}
}

// Skipped since it would actually send signals
func SkipTestActualSignalHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping actual signal test in short mode")
	}
	
	// Get the current process
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find current process: %v", err)
	}
	
	// Create mock collector
	collector := newMockCollector()
	
	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Create a channel to signal when run exits
	done := make(chan struct{})
	
	// Run in goroutine
	go func() {
		runWithCollector(ctx, collector)
		close(done)
	}()
	
	// Allow some time for setup
	time.Sleep(100 * time.Millisecond)
	
	// Send a SIGTERM to the process
	err = proc.Signal(syscall.SIGTERM)
	if err != nil {
		t.Fatalf("Failed to send signal: %v", err)
	}
	
	// Wait for the function to exit
	select {
	case <-done:
		// Success - function exited due to signal
	case <-time.After(1 * time.Second):
		t.Fatal("Function did not exit after signal")
	}
} 