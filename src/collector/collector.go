package main

import (
	"context"
	"fmt"
)

// CollectorError represents an error from a collector
type CollectorError struct {
	Message     string
	Recoverable bool
}

func (e *CollectorError) Error() string {
	return e.Message
}

// Collector interface for collecting events
type Collector interface {
	Run(ctx context.Context) error
	Stop()
}

// CollectorFactory is a function type for creating collectors
type CollectorFactory func(config *Config, processor EventProcessor) (Collector, error)

// CollectorManager manages the lifecycle of collectors
type CollectorManager struct {
	config         *Config
	processor      EventProcessor
	createCollector CollectorFactory
	currentCollector Collector
}

// EventProcessor interface for processing events
type EventProcessor interface {
	ProcessEvent(Event)
	Run(ctx context.Context)
}

// NewCollectorManager creates a new collector manager
func NewCollectorManager(config *Config, processor EventProcessor, factory CollectorFactory) (*CollectorManager, error) {
	manager := &CollectorManager{
		config:         config,
		processor:      processor,
		createCollector: factory,
	}
	
	// Create initial collector
	collector, err := factory(config, processor)
	if err != nil {
		return nil, fmt.Errorf("failed to create collector: %w", err)
	}
	
	manager.currentCollector = collector
	return manager, nil
}

// Run starts the collector manager
func (m *CollectorManager) Run(ctx context.Context) {
	// Start the event processor
	processorCtx, processorCancel := context.WithCancel(ctx)
	defer processorCancel()
	
	go m.processor.Run(processorCtx)
	
	// Main collector loop with restart capability
	for {
		// Check if context is done
		select {
		case <-ctx.Done():
			if m.currentCollector != nil {
				m.currentCollector.Stop()
			}
			return
		default:
			// Continue running
		}
		
		// Run the current collector
		err := m.currentCollector.Run(ctx)
		
		// If context is done, exit gracefully
		select {
		case <-ctx.Done():
			return
		default:
			// Continue to potential restart
		}
		
		// Handle collector errors
		if err != nil {
			collectorErr, ok := err.(*CollectorError)
			if ok && collectorErr.Recoverable {
				// Create a new collector
				newCollector, createErr := m.createCollector(m.config, m.processor)
				if createErr != nil {
					// Failed to create new collector, exit
					return
				}
				
				// Replace the current collector
				m.currentCollector = newCollector
			} else {
				// Non-recoverable error, exit
				return
			}
		} else {
			// Collector exited without error, exit manager
			return
		}
	}
} 