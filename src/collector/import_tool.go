package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ImportTool is a simple command-line tool to import event data from backup files to the database
func ImportTool() {
	// Only run if we're specifically using the import command
	if len(os.Args) < 2 || os.Args[1] != "import" {
		return
	}

	// Configure import flags
	importCmd := flag.NewFlagSet("import", flag.ExitOnError)
	configPath := importCmd.String("config", "config.yaml", "Path to configuration file")
	filePath := importCmd.String("file", "", "Path to event file to import (required)")
	dirPath := importCmd.String("dir", "", "Directory containing event files to import")
	batchSize := importCmd.Int("batch", 100, "Batch size for importing events")

	// Parse import flags (skip the "import" arg)
	if err := importCmd.Parse(os.Args[2:]); err != nil {
		log.Fatalf("Error parsing import flags: %v", err)
	}

	// Validate required parameters
	if *filePath == "" && *dirPath == "" {
		log.Fatalf("Error: Either -file or -dir must be specified")
	}

	// Load configuration
	flag.Set("config", *configPath) // Set the config flag for loadConfig
	cfg := loadConfig()

	// Ensure PostgreSQL is enabled in the config
	if !cfg.PostgresEnabled {
		log.Fatalf("Error: PostgreSQL must be enabled in the configuration for import")
	}

	// Create PostgreSQL client
	dbClient, err := NewPostgresClient(cfg)
	if err != nil {
		log.Fatalf("Error connecting to PostgreSQL: %v", err)
	}
	defer dbClient.Close()

	// Process a single file
	if *filePath != "" {
		if err := ImportEventsFromFile(*filePath, dbClient, *batchSize); err != nil {
			log.Fatalf("Error importing from file %s: %v", *filePath, err)
		}
		fmt.Println("Import completed successfully")
		os.Exit(0)
	}

	// Process a directory of files
	if *dirPath != "" {
		files, err := filepath.Glob(filepath.Join(*dirPath, "events_*.jsonl"))
		if err != nil {
			log.Fatalf("Error finding event files: %v", err)
		}

		if len(files) == 0 {
			log.Fatalf("No event files found in directory %s", *dirPath)
		}

		fmt.Printf("Found %d event files to import\n", len(files))
		
		// Sort files by name (which includes timestamp)
		// This isn't a perfect sort by date, but works for the filename format we use
		// events_YYYYMMDD_HHMMSS.jsonl
		sort.Strings(files)
		
		for i, file := range files {
			fmt.Printf("[%d/%d] Importing file: %s\n", i+1, len(files), filepath.Base(file))
			if err := ImportEventsFromFile(file, dbClient, *batchSize); err != nil {
				log.Printf("Error importing from file %s: %v", file, err)
				// Continue with the next file
			}
		}
		
		fmt.Println("Import completed successfully")
		os.Exit(0)
	}
} 