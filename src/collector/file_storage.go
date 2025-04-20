package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileBackupClient implements the DBClient interface for storing events to files
type FileBackupClient struct {
	basePath     string
	currentFile  *os.File
	mu           sync.Mutex
	enabled      bool
	currentBatch int
	maxFileSize  int64
	maxFileAge   time.Duration
	fileCreated  time.Time
	fileSize     int64
}

// FileBackupConfig contains configuration for file backup
type FileBackupConfig struct {
	Enabled     bool
	BasePath    string
	MaxFileSize int64        // Maximum file size in bytes before rotation
	MaxFileAge  time.Duration // Maximum file age before rotation
}

// NewFileBackupClient creates a new client for storing events to files
func NewFileBackupClient(config FileBackupConfig) (DBClient, error) {
	if !config.Enabled {
		return nil, nil
	}

	// Set defaults if not specified
	if config.MaxFileSize == 0 {
		config.MaxFileSize = 10 * 1024 * 1024 // 10MB default
	}
	if config.MaxFileAge == 0 {
		config.MaxFileAge = 1 * time.Hour // 1 hour default
	}
	if config.BasePath == "" {
		config.BasePath = "events"
	}

	// Ensure base directory exists
	if err := os.MkdirAll(config.BasePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	client := &FileBackupClient{
		basePath:    config.BasePath,
		enabled:     true,
		maxFileSize: config.MaxFileSize,
		maxFileAge:  config.MaxFileAge,
	}

	log.Printf("File backup initialized. Base path: %s", config.BasePath)
	return client, nil
}

// StoreEvents writes events to a JSON file
func (f *FileBackupClient) StoreEvents(events []Event) error {
	if !f.enabled || len(events) == 0 {
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// Check if we need to open a new file
	if f.currentFile == nil || f.shouldRotateFile() {
		if err := f.rotateFile(); err != nil {
			return fmt.Errorf("failed to rotate file: %w", err)
		}
	}

	// Ensure the file is open
	if f.currentFile == nil {
		return fmt.Errorf("file is not open for writing")
	}

	// Serialize events as JSON lines
	for _, event := range events {
		// Create a record with timestamp
		record := struct {
			Timestamp time.Time `json:"timestamp"`
			Type      string    `json:"type"`
			Data      json.RawMessage `json:"data"`
		}{
			Timestamp: time.Now(),
			Type:      event.Type,
			Data:      event.Data,
		}

		data, err := json.Marshal(record)
		if err != nil {
			log.Printf("Error marshaling event to JSON: %v", err)
			continue
		}

		// Write the JSON line with a newline separator
		data = append(data, '\n')
		n, err := f.currentFile.Write(data)
		if err != nil {
			return fmt.Errorf("failed to write event to file: %w", err)
		}

		f.fileSize += int64(n)
	}

	// Flush to disk
	if err := f.currentFile.Sync(); err != nil {
		log.Printf("Warning: Failed to sync file: %v", err)
	}

	log.Printf("Stored %d events in backup file", len(events))
	return nil
}

// shouldRotateFile checks if the current file should be rotated
func (f *FileBackupClient) shouldRotateFile() bool {
	if f.currentFile == nil {
		return true
	}

	// Check file size
	if f.fileSize >= f.maxFileSize {
		return true
	}

	// Check file age
	if time.Since(f.fileCreated) >= f.maxFileAge {
		return true
	}

	return false
}

// rotateFile closes the current file and opens a new one
func (f *FileBackupClient) rotateFile() error {
	// Close current file if open
	if f.currentFile != nil {
		if err := f.currentFile.Close(); err != nil {
			log.Printf("Warning: Failed to close file: %v", err)
		}
		f.currentFile = nil
	}

	// Create a new file with timestamp in the name
	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(f.basePath, fmt.Sprintf("events_%s.jsonl", timestamp))
	
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}

	f.currentFile = file
	f.fileCreated = time.Now()
	f.fileSize = 0
	f.currentBatch++

	log.Printf("Created new backup file: %s", filename)
	return nil
}

// ImportEventsFromFile imports events from a specified backup file
func ImportEventsFromFile(filePath string, dbClient DBClient, batchSize int) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	
	batch := make([]Event, 0, batchSize)
	recordCount := 0
	
	// Process one JSON object at a time
	for decoder.More() {
		var record struct {
			Timestamp time.Time `json:"timestamp"`
			Type      string    `json:"type"`
			Data      json.RawMessage `json:"data"`
		}
		
		if err := decoder.Decode(&record); err != nil {
			log.Printf("Error decoding record: %v", err)
			continue
		}
		
		// Convert to Event format
		event := Event{
			Type: record.Type,
			Data: record.Data,
		}
		
		batch = append(batch, event)
		recordCount++
		
		// Process in batches
		if len(batch) >= batchSize {
			if err := dbClient.StoreEvents(batch); err != nil {
				log.Printf("Error storing events from file: %v", err)
			}
			batch = batch[:0] // Clear batch but keep capacity
		}
	}
	
	// Process any remaining events
	if len(batch) > 0 {
		if err := dbClient.StoreEvents(batch); err != nil {
			log.Printf("Error storing final batch of events from file: %v", err)
		}
	}
	
	log.Printf("Imported %d events from %s", recordCount, filePath)
	return nil
}

// Close closes the current file
func (f *FileBackupClient) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.currentFile != nil {
		err := f.currentFile.Close()
		f.currentFile = nil
		return err
	}
	return nil
}

// GetMetrics returns metrics about the file backup client
func (f *FileBackupClient) GetMetrics() map[string]interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	metrics := map[string]interface{}{
		"type":               "file_backup",
		"enabled":            f.enabled,
		"base_path":          f.basePath,
		"current_batch":      f.currentBatch,
		"max_file_size":      f.maxFileSize,
		"max_file_age_hours": f.maxFileAge.Hours(),
	}

	if f.currentFile != nil {
		metrics["current_file_open"] = true
		metrics["current_file_size"] = f.fileSize
		metrics["file_created_at"] = f.fileCreated.Format(time.RFC3339)
		metrics["file_age_minutes"] = time.Since(f.fileCreated).Minutes()
	} else {
		metrics["current_file_open"] = false
	}

	return metrics
} 