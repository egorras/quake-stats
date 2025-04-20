package main

import (
	"fmt"
	"log"
	"sync"
)

// MultiDBClient implements the DBClient interface and stores events in multiple storage backends
type MultiDBClient struct {
	clients []DBClient
}

// NewMultiDBClient creates a new multi-client for storing events in multiple backends
func NewMultiDBClient(clients []DBClient) DBClient {
	return &MultiDBClient{
		clients: clients,
	}
}

// StoreEvents stores events in all clients
func (m *MultiDBClient) StoreEvents(events []Event) error {
	if len(events) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errs := make([]error, len(m.clients))

	// Store events in all clients in parallel
	for i, client := range m.clients {
		wg.Add(1)
		go func(idx int, c DBClient) {
			defer wg.Done()
			if err := c.StoreEvents(events); err != nil {
				errs[idx] = err
				log.Printf("Error storing events in client %d: %v", idx, err)
			}
		}(i, client)
	}

	// Wait for all storage operations to complete
	wg.Wait()

	// Check for errors
	var errorMsgs string
	errorCount := 0
	for i, err := range errs {
		if err != nil {
			errorCount++
			errorMsgs += fmt.Sprintf("Client %d: %v; ", i, err)
		}
	}

	// If all clients failed, return an error
	if errorCount == len(m.clients) {
		return fmt.Errorf("all storage clients failed: %s", errorMsgs)
	}

	// If some clients failed but at least one succeeded, just log the errors
	if errorCount > 0 {
		log.Printf("Warning: %d/%d storage clients had errors: %s", 
			errorCount, len(m.clients), errorMsgs)
	}

	return nil
}

// Close closes all clients
func (m *MultiDBClient) Close() error {
	var lastErr error
	for i, client := range m.clients {
		if err := client.Close(); err != nil {
			log.Printf("Error closing client %d: %v", i, err)
			lastErr = err
		}
	}
	return lastErr
}

// GetMetrics returns combined metrics from all clients
func (m *MultiDBClient) GetMetrics() map[string]interface{} {
	metrics := map[string]interface{}{
		"client_count": len(m.clients),
	}
	
	// Collect metrics from each client
	for i, client := range m.clients {
		clientMetrics := client.GetMetrics()
		if clientMetrics != nil {
			prefix := fmt.Sprintf("client_%d_", i)
			for k, v := range clientMetrics {
				metrics[prefix+k] = v
			}
		}
	}
	
	return metrics
} 