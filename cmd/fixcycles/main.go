package main

import (
	"encoding/json"
	"fmt"
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/variable"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Get project root
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	// Find project root (where go.mod is)
	projectRoot := wd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			fmt.Println("Could not find project root (go.mod)")
			os.Exit(1)
		}
		projectRoot = parent
	}

	dataPath := filepath.Join(projectRoot, "archive", "Data", "json", "hash_data.json")

	// Read the data file
	data, err := os.ReadFile(dataPath)
	if err != nil {
		fmt.Printf("Error reading hash_data.json: %v\n", err)
		os.Exit(1)
	}

	var hData map[string]m.Hash
	if err := json.Unmarshal(data, &hData); err != nil {
		fmt.Printf("Error unmarshaling JSON: %v\n", err)
		os.Exit(1)
	}

	// Create encoder
	varReg := variable.VarReg
	encoder, err := node.NewEncoder(varReg)
	if err != nil {
		fmt.Printf("Error creating encoder: %v\n", err)
		os.Exit(1)
	}

	// Find containers to delete (those involved in cycles)
	containersToDelete := findContainersInCycles(hData, encoder)

	if len(containersToDelete) == 0 {
		fmt.Println("✓ No cycles found - nothing to fix")
		return
	}

	fmt.Printf("Found %d container(s) to delete:\n", len(containersToDelete))
	for id := range containersToDelete {
		fmt.Printf("  - %s\n", id)
	}

	// Delete the containers
	for containerID := range containersToDelete {
		deleteContainer(hData, encoder, containerID)
		fmt.Printf("✓ Deleted container %s\n", containerID)
	}

	// Save the fixed data
	if err := saveJSON(dataPath, hData); err != nil {
		fmt.Printf("Error saving JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Fixed data saved to %s\n", dataPath)
	fmt.Println("You can now restart your application and the timer plugin should register correctly.")
}

// findContainersInCycles detects cycles and returns the container IDs involved
func findContainersInCycles(hData map[string]m.Hash, encoder node.Encoder) map[string]bool {
	containersToDelete := make(map[string]bool)

	// Build container mappings
	containerTriggers := make(map[string]m.Hash)
	containerResults := make(map[string]m.Hash)

	for key, value := range hData {
		if strings.HasPrefix(key, "C->T:") {
			containerID := strings.TrimPrefix(key, "C->T:")
			containerTriggers[containerID] = value
		} else if strings.HasPrefix(key, "C->R:") {
			containerID := strings.TrimPrefix(key, "C->R:")
			containerResults[containerID] = value
		}
	}

	// Check each container for cycles
	for containerID, triggers := range containerTriggers {
		results := containerResults[containerID]
		if results == nil {
			continue
		}

		// For each result, check if it triggers any container that eventually produces a trigger
		for resultNode := range results {
			encodedResult := encoder.Do(resultNode)
			triggerKey := "T->C:" + encodedResult
			if containers, ok := hData[triggerKey]; ok {
				for triggeredContainerID := range containers {
					if triggeredContainerID == containerID {
						// Self-reference
						containersToDelete[containerID] = true
						continue
					}

					// Check if the triggered container produces any of the original triggers
					triggeredResultsKey := "C->R:" + triggeredContainerID
					if triggeredResults, ok := hData[triggeredResultsKey]; ok {
						for triggeredResult := range triggeredResults {
							encodedTriggeredResult := encoder.Do(triggeredResult)
							for triggerNode := range triggers {
								encodedTrigger := encoder.Do(triggerNode)
								if encodedTriggeredResult == encodedTrigger {
									// Cycle detected!
									containersToDelete[containerID] = true
									containersToDelete[triggeredContainerID] = true
								}
							}
						}
					}
				}
			}
		}
	}

	return containersToDelete
}

// deleteContainer removes a container and all its related entries from the data
func deleteContainer(hData map[string]m.Hash, encoder node.Encoder, containerID string) {
	// Get the container's triggers and results
	tTag := "C->T:" + containerID
	rTag := "C->R:" + containerID

	triggers := hData[tTag]
	results := hData[rTag]

	// Remove T->C mappings (trigger -> container)
	for trigger := range triggers {
		encodedTrigger := encoder.Do(trigger)
		t2cTag := "T->C:" + encodedTrigger
		if containers, ok := hData[t2cTag]; ok {
			delete(containers, containerID)
			if len(containers) == 0 {
				delete(hData, t2cTag)
			} else {
				hData[t2cTag] = containers
			}
		}
	}

	// Remove R->C mappings (result -> container)
	for result := range results {
		encodedResult := encoder.Do(result)
		r2cTag := "R->C:" + encodedResult
		if containers, ok := hData[r2cTag]; ok {
			delete(containers, containerID)
			if len(containers) == 0 {
				delete(hData, r2cTag)
			} else {
				hData[r2cTag] = containers
			}
		}
	}

	// Remove C->T and C->R mappings
	delete(hData, tTag)
	delete(hData, rTag)
}

// saveJSON saves the data back to the JSON file
func saveJSON(path string, data map[string]m.Hash) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, jsonData, 0644)
}
