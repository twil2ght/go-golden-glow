package main

import (
	"encoding/json"
	"fmt"
	"goldenglow/components/runner"
	"goldenglow/config"
	"goldenglow/node"
	"goldenglow/setup"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: jsoninput <path-to-json-file-or-folder>")
		fmt.Println("  jsoninput <file.json>     - Process single JSON file")
		fmt.Println("  jsoninput <folder>        - Process all JSON files in folder (builder.json first)")
		os.Exit(1)
	}

	inputPath := os.Args[1]

	setup.Init()

	// Collect all inputs
	inputs := collectInputs(inputPath)
	fmt.Printf("Total inputs to process: %d\n", len(inputs))

	// Process each input directly
	r := runner.DefaultRunner()
	for i, input := range inputs {
		// Preprocess: "Zero says <input> to Susie"
		preprocessed := fmt.Sprintf("%s says %s to %s", config.User, input, config.GG)
		fmt.Printf("\n=== Processing input %d/%d ===\n", i+1, len(inputs))
		fmt.Printf("Original: %s\n", input)
		fmt.Printf("Preprocessed: %s\n", preprocessed)

		// Create node and run directly
		item, err := node.DefaultFactory().New(preprocessed)
		if err != nil {
			fmt.Printf("Error creating node: %v\n", err)
			continue
		}

		if err := r.Run(item); err != nil {
			fmt.Printf("Error running item: %v\n", err)
		}
	}

	fmt.Println("\n=== All inputs processed ===")
	setup.Shutdown()
}

func collectInputs(inputPath string) []string {
	var inputs []string

	info, err := os.Stat(inputPath)
	if err != nil {
		fmt.Printf("Error accessing path: %v\n", err)
		return inputs
	}

	if info.IsDir() {
		inputs = collectFromDirectory(inputPath)
	} else {
		inputs = collectFromFile(inputPath)
	}

	return inputs
}

func collectFromDirectory(dirPath string) []string {
	var allInputs []string

	// First, process builder.json if it exists
	builderPath := filepath.Join(dirPath, "builder.json")
	if _, err := os.Stat(builderPath); err == nil {
		fmt.Println("Processing builder.json first...")
		allInputs = append(allInputs, collectFromFile(builderPath)...)
	}

	// Then process all other JSON files
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return allInputs
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subDirPath := filepath.Join(dirPath, entry.Name())
			allInputs = append(allInputs, collectFromSubDirectory(subDirPath)...)
		} else if entry.Name() != "builder.json" && filepath.Ext(entry.Name()) == ".json" {
			filePath := filepath.Join(dirPath, entry.Name())
			allInputs = append(allInputs, collectFromFile(filePath)...)
		}
	}

	return allInputs
}

func collectFromSubDirectory(dirPath string) []string {
	var allInputs []string

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("Error reading subdirectory: %v\n", err)
		return allInputs
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			filePath := filepath.Join(dirPath, entry.Name())
			allInputs = append(allInputs, collectFromFile(filePath)...)
		}
	}

	return allInputs
}

func collectFromFile(filePath string) []string {
	var inputs []string

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filePath, err)
		return inputs
	}

	// Try to parse as array of objects first
	var interfaceArray []struct {
		Inputs []string `json:"inputs"`
	}

	err = json.Unmarshal(data, &interfaceArray)
	if err == nil {
		for _, item := range interfaceArray {
			for _, input := range item.Inputs {
				if input != "exit" && input != "quit" {
					inputs = append(inputs, input)
				}
			}
		}
		return inputs
	}

	// Try to parse as single object with inputs
	var singleData struct {
		Inputs []string `json:"inputs"`
	}

	err = json.Unmarshal(data, &singleData)
	if err != nil {
		fmt.Printf("Error parsing file %s: %v\n", filePath, err)
		return inputs
	}

	for _, input := range singleData.Inputs {
		if input != "exit" && input != "quit" {
			inputs = append(inputs, input)
		}
	}

	return inputs
}
