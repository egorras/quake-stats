package main

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	// Test default configuration
	config, err := loadConfig("config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check default values
	if config.ZMQ.Endpoint == "" {
		t.Error("ZMQ endpoint should have a default value")
	}

	if config.Logging.Level == "" {
		t.Error("Logging level should have a default value")
	}

	if config.Events.FlushInterval == 0 {
		t.Error("Events flush interval should have a default value")
	}

	if config.Events.FlushSize == 0 {
		t.Error("Events flush size should have a default value")
	}
}

func TestConfigEnvOverrides(t *testing.T) {
	// Set environment variables to override config
	os.Setenv("ZMQ_ENDPOINT", "tcp://test-server:5555")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("FLUSH_INTERVAL", "5")
	os.Setenv("FLUSH_SIZE", "100")

	// Clean up environment after the test
	defer func() {
		os.Unsetenv("ZMQ_ENDPOINT")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("FLUSH_INTERVAL")
		os.Unsetenv("FLUSH_SIZE")
	}()

	config, err := loadConfig("config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check if environment variables override config values
	if config.ZMQ.Endpoint != "tcp://test-server:5555" {
		t.Errorf("ZMQ endpoint not overridden by environment variable, got %s", config.ZMQ.Endpoint)
	}

	if config.Logging.Level != "debug" {
		t.Errorf("Logging level not overridden by environment variable, got %s", config.Logging.Level)
	}

	if config.Events.FlushInterval != 5 {
		t.Errorf("Events flush interval not overridden by environment variable, got %d", config.Events.FlushInterval)
	}

	if config.Events.FlushSize != 100 {
		t.Errorf("Events flush size not overridden by environment variable, got %d", config.Events.FlushSize)
	}
} 