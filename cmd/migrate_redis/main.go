package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// expirableHash matches storage/redis.go structure: map[value]expiration_time
type expirableHash map[string]time.Time

func main() {
	// Define paths
	archiveDir := filepath.Join("archive", "Data", "redis")
	sourceFile := filepath.Join(archiveDir, "data.json")
	targetFile := filepath.Join(archiveDir, "hash_data.json")

	// Read source file
	fmt.Printf("Reading from: %s\n", sourceFile)
	sourceData, err := os.ReadFile(sourceFile)
	if err != nil {
		fmt.Printf("Error reading source file: %v\n", err)
		os.Exit(1)
	}

	// Parse source JSON (old data.json format: key -> {value, expiration})
	var sourceMap map[string]struct {
		Value      string    `json:"value"`
		Expiration time.Time `json:"expiration"`
	}
	if err := json.Unmarshal(sourceData, &sourceMap); err != nil {
		fmt.Printf("Error parsing source JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d entries to migrate\n", len(sourceMap))

	// Convert to hash_data.json format: key -> {value -> expiration}
	targetMap := make(map[string]expirableHash)
	for key, entry := range sourceMap {
		// Create expirableHash where value maps to expiration time
		hash := make(expirableHash)
		hash[entry.Value] = entry.Expiration
		targetMap[key] = hash
	}

	// Write to target file
	output, err := json.MarshalIndent(targetMap, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling target JSON: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(targetFile, output, 0644); err != nil {
		fmt.Printf("Error writing target file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully migrated to: %s\n", targetFile)
	fmt.Println("Migration complete!")
}
