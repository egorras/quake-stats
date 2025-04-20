package main

import (
	"flag"
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all the configuration parameters
type Config struct {
	ZmqEndpoint             string
	BatchSize               int
	FlushIntervalSec        int
	VerboseLogging          bool
	PostgresEnabled         bool
	PostgresConnectionString string
	PostgresTable           string
	PostgresIdleTimeoutMin  int
	FileBackupEnabled       bool
	FileBackupPath          string
	FileBackupMaxSizeMB     int
	FileBackupMaxAgeHours   int
}

// configFlagDefined tracks if the config flag has been defined already
var configFlagDefined bool

// loadConfig reads configuration from file and environment variables
func loadConfig() Config {
	var configPath string
	
	// Only define the flag if it hasn't been defined yet
	if !configFlagDefined {
		flag.StringVar(&configPath, "config", "config.yaml", "Path to configuration file (default: config.yaml)")
		configFlagDefined = true
	}
	
	// Parse flags only if they haven't been parsed already
	if !flag.Parsed() {
		flag.Parse()
	}

	// Initialize viper
	v := viper.New()

	// Set default values
	v.SetDefault("zmq_endpoint", "tcp://89.168.29.137:27960")
	v.SetDefault("batch_size", 10)
	v.SetDefault("flush_interval_sec", 1)
	v.SetDefault("verbose_logging", true)
	
	// PostgreSQL defaults
	v.SetDefault("postgres_enabled", false)
	v.SetDefault("postgres_connection_string", "postgresql://postgres:postgres@localhost:5432/quake_stats?sslmode=disable")
	v.SetDefault("postgres_table", "events")
	v.SetDefault("postgres_idle_timeout_min", 5)
	
	// File backup defaults
	v.SetDefault("file_backup_enabled", true)
	v.SetDefault("file_backup_path", "backup/events")
	v.SetDefault("file_backup_max_size_mb", 10)
	v.SetDefault("file_backup_max_age_hours", 1)

	// Configure viper to read environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Configure file-based config
	if configPath != "" {
		// Set config type based on file extension
		ext := filepath.Ext(configPath)
		if ext != "" {
			v.SetConfigType(ext[1:]) // Remove the leading dot
		}

		v.SetConfigFile(configPath)
		err := v.ReadInConfig()
		if err != nil {
			log.Printf("Warning: Using default configuration values: %v", err)
		} else {
			log.Printf("Loaded configuration from %s", configPath)
		}
	}

	// Create and return the configuration
	config := Config{
		ZmqEndpoint:             v.GetString("zmq_endpoint"),
		BatchSize:               v.GetInt("batch_size"),
		FlushIntervalSec:        v.GetInt("flush_interval_sec"),
		VerboseLogging:          v.GetBool("verbose_logging"),
		PostgresEnabled:         v.GetBool("postgres_enabled"),
		PostgresConnectionString: v.GetString("postgres_connection_string"),
		PostgresTable:           v.GetString("postgres_table"),
		PostgresIdleTimeoutMin:  v.GetInt("postgres_idle_timeout_min"),
		FileBackupEnabled:       v.GetBool("file_backup_enabled"),
		FileBackupPath:          v.GetString("file_backup_path"),
		FileBackupMaxSizeMB:     v.GetInt("file_backup_max_size_mb"),
		FileBackupMaxAgeHours:   v.GetInt("file_backup_max_age_hours"),
	}

	return config
}

// logConfig logs the current configuration
func logConfig(cfg Config) {
	log.Printf("Starting collector with configuration:")
	log.Printf("- ZMQ Endpoint: %s", cfg.ZmqEndpoint)
	log.Printf("- Batch Size: %d", cfg.BatchSize)
	log.Printf("- Flush Interval: %d seconds", cfg.FlushIntervalSec)
	log.Printf("- Verbose Logging: %v", cfg.VerboseLogging)
	log.Printf("- Postgres Enabled: %v", cfg.PostgresEnabled)
	if cfg.PostgresEnabled {
		log.Printf("- Postgres: Using connection string")
		log.Printf("- Postgres Table: %s", cfg.PostgresTable)
		log.Printf("- Postgres Idle Timeout: %d minutes", cfg.PostgresIdleTimeoutMin)
	}
	log.Printf("- File Backup Enabled: %v", cfg.FileBackupEnabled)
	if cfg.FileBackupEnabled {
		log.Printf("- File Backup Path: %s", cfg.FileBackupPath)
		log.Printf("- File Backup Max Size: %d MB", cfg.FileBackupMaxSizeMB)
		log.Printf("- File Backup Max Age: %d hours", cfg.FileBackupMaxAgeHours)
	}
} 