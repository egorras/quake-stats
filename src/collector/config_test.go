package main

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	// Test default configuration
	os.Args = []string{"collector"}
	config := loadConfig()

	// Check default values
	if config.ZmqEndpoint == "" {
		t.Error("ZMQ endpoint should have a default value")
	}

	if config.BatchSize <= 0 {
		t.Error("Batch size should have a default value")
	}

	if config.FlushIntervalSec <= 0 {
		t.Error("Flush interval should have a default value")
	}
}

func TestConfigEnvOverrides(t *testing.T) {
	// Set environment variables to override config
	os.Setenv("ZMQ_ENDPOINT", "tcp://test-server:5555")
	os.Setenv("BATCH_SIZE", "50")
	os.Setenv("FLUSH_INTERVAL_SEC", "5")
	os.Setenv("VERBOSE_LOGGING", "false")
	os.Setenv("POSTGRES_ENABLED", "true")
	os.Setenv("POSTGRES_CONNECTION_STRING", "postgresql://test:test@testdb:5432/test")
	os.Setenv("POSTGRES_TABLE", "test_events")

	// Clean up environment after the test
	defer func() {
		os.Unsetenv("ZMQ_ENDPOINT")
		os.Unsetenv("BATCH_SIZE")
		os.Unsetenv("FLUSH_INTERVAL_SEC")
		os.Unsetenv("VERBOSE_LOGGING")
		os.Unsetenv("POSTGRES_ENABLED")
		os.Unsetenv("POSTGRES_CONNECTION_STRING")
		os.Unsetenv("POSTGRES_TABLE")
	}()

	os.Args = []string{"collector"}
	config := loadConfig()

	// Check if environment variables override config values
	if config.ZmqEndpoint != "tcp://test-server:5555" {
		t.Errorf("ZMQ endpoint not overridden by environment variable, got %s", config.ZmqEndpoint)
	}

	if config.BatchSize != 50 {
		t.Errorf("Batch size not overridden by environment variable, got %d", config.BatchSize)
	}

	if config.FlushIntervalSec != 5 {
		t.Errorf("Flush interval not overridden by environment variable, got %d", config.FlushIntervalSec)
	}
	
	if config.VerboseLogging != false {
		t.Errorf("Verbose logging not overridden by environment variable, got %v", config.VerboseLogging)
	}
	
	if config.PostgresEnabled != true {
		t.Errorf("Postgres enabled not overridden by environment variable, got %v", config.PostgresEnabled)
	}
	
	if config.PostgresConnectionString != "postgresql://test:test@testdb:5432/test" {
		t.Errorf("Postgres connection string not overridden by environment variable, got %s", config.PostgresConnectionString)
	}
	
	if config.PostgresTable != "test_events" {
		t.Errorf("Postgres table not overridden by environment variable, got %s", config.PostgresTable)
	}
} 