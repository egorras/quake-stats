package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	zmq4 "github.com/pebbe/zmq4"
)

// ZmqCollector represents a ZMQ event collector
type ZmqCollector struct {
	endpoint   string
	socket     *zmq4.Socket
	processor  EventProcessorInterface
	isRemote   bool
	cancelFunc context.CancelFunc
}

// NewZmqCollector creates a new ZMQ collector
func NewZmqCollector(endpoint string, processor EventProcessorInterface) (*ZmqCollector, error) {
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

// Run implements the Collector interface
func (c *ZmqCollector) Run(ctx context.Context) error {
	// Create a new context that we can cancel if Stop is called
	runCtx, cancel := context.WithCancel(ctx)
	c.cancelFunc = cancel
	
	// Start the collector and return any error
	return c.Start(runCtx)
}

// Stop implements the Collector interface
func (c *ZmqCollector) Stop() {
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
	c.Close()
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

		c.processor.ProcessEvent(e)
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