package container

import (
	"encoding/json"
	"fmt"
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/storage"
	"goldenglow/variable"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFindAllCyclesInData loads the actual hash_data.json and detects all cycles
func TestFindAllCyclesInData(t *testing.T) {
	// Load the actual data file
	dataPath := filepath.Join(getProjectRoot(), "archive", "Data", "json", "hash_data.json")

	data, err := os.ReadFile(dataPath)
	if err != nil {
		t.Fatalf("Failed to read hash_data.json: %v", err)
	}

	var hData map[string]m.Hash
	if err := json.Unmarshal(data, &hData); err != nil {
		t.Fatalf("Failed to unmarshal hash_data.json: %v", err)
	}

	// Create a test repository with the loaded data
	repo := &testRepository{data: hData}

	// Create encoder
	varReg := variable.VarReg
	encoder, err := node.NewEncoder(varReg)
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}

	// Create cycle detector
	detector := NewCycleDetector(repo, encoder)

	// Find all cycles
	cycles, err := detector.DetectExistingCycles()
	if err != nil {
		t.Fatalf("Failed to detect cycles: %v", err)
	}

	// Report results
	if len(cycles) == 0 {
		t.Log("✓ No cycles detected in the data")
	} else {
		t.Logf("⚠ Found %d cycle(s) in the data:", len(cycles))
		for i, cycle := range cycles {
			t.Logf("\nCycle %d:", i+1)
			for _, step := range cycle.Path {
				t.Logf("  → %s", step)
			}
		}
	}
}

// TestFindCyclesDetailed performs a detailed analysis of potential cycles
func TestFindCyclesDetailed(t *testing.T) {
	dataPath := filepath.Join(getProjectRoot(), "archive", "Data", "json", "hash_data.json")

	data, err := os.ReadFile(dataPath)
	if err != nil {
		t.Fatalf("Failed to read hash_data.json: %v", err)
	}

	var hData map[string]m.Hash
	if err := json.Unmarshal(data, &hData); err != nil {
		t.Fatalf("Failed to unmarshal hash_data.json: %v", err)
	}

	// Analyze container relationships
	containerTriggers := make(map[string]m.Hash) // containerID -> trigger nodes
	containerResults := make(map[string]m.Hash)  // containerID -> result nodes

	for key, value := range hData {
		if strings.HasPrefix(key, "C->T:") {
			containerID := strings.TrimPrefix(key, "C->T:")
			containerTriggers[containerID] = value
		} else if strings.HasPrefix(key, "C->R:") {
			containerID := strings.TrimPrefix(key, "C->R:")
			containerResults[containerID] = value
		}
	}

	t.Logf("Found %d containers", len(containerTriggers))

	// Check each container for potential cycles
	var foundCycles []string
	for containerID, triggers := range containerTriggers {
		results := containerResults[containerID]
		if results == nil {
			continue
		}

		// For each result, check if it triggers any container that eventually produces a trigger
		for resultNode := range results {
			triggerKey := "T->C:" + encodeNode(resultNode)
			if containers, ok := hData[triggerKey]; ok {
				for triggeredContainerID := range containers {
					if triggeredContainerID == containerID {
						// Self-reference detected
						foundCycles = append(foundCycles,
							fmt.Sprintf("Container %s: result[%s] triggers itself", containerID, resultNode))
						continue
					}

					// Check if the triggered container produces any of the original triggers
					triggeredResultsKey := "C->R:" + triggeredContainerID
					if triggeredResults, ok := hData[triggeredResultsKey]; ok {
						for triggeredResult := range triggeredResults {
							for triggerNode := range triggers {
								if encodeNode(triggeredResult) == encodeNode(triggerNode) {
									foundCycles = append(foundCycles,
										fmt.Sprintf("Cycle: Container[%s] -> result[%s] -> triggers Container[%s] -> result[%s] (matches trigger[%s])",
											containerID, resultNode, triggeredContainerID, triggeredResult, triggerNode))
								}
							}
						}
					}
				}
			}
		}
	}

	if len(foundCycles) == 0 {
		t.Log("✓ No cycles found in detailed analysis")
	} else {
		t.Logf("⚠ Found %d potential cycle(s):", len(foundCycles))
		for _, cycle := range foundCycles {
			t.Logf("  - %s", cycle)
		}
	}
}

// TestContainerRelationships prints all container relationships for manual inspection
func TestContainerRelationships(t *testing.T) {
	dataPath := filepath.Join(getProjectRoot(), "archive", "Data", "json", "hash_data.json")

	data, err := os.ReadFile(dataPath)
	if err != nil {
		t.Fatalf("Failed to read hash_data.json: %v", err)
	}

	var hData map[string]m.Hash
	if err := json.Unmarshal(data, &hData); err != nil {
		t.Fatalf("Failed to unmarshal hash_data.json: %v", err)
	}

	// Build trigger->results mapping for each container
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

	t.Log("Container Relationships:")
	t.Log("========================")

	for containerID := range containerTriggers {
		triggers := containerTriggers[containerID]
		results := containerResults[containerID]

		t.Logf("\nContainer: %s", containerID)
		t.Log("  Triggers:")
		for trigger := range triggers {
			t.Logf("    - %s", trigger)
		}
		t.Log("  Results:")
		for result := range results {
			t.Logf("    - %s", result)
		}

		// Show what containers each result triggers
		for result := range results {
			triggerKey := "T->C:" + encodeNode(result)
			if containers, ok := hData[triggerKey]; ok && len(containers) > 0 {
				t.Logf("  Result '%s' triggers containers:", result)
				for cid := range containers {
					t.Logf("    → %s", cid)
				}
			}
		}
	}
}

// Helper function to encode node value (simplified version)
func encodeNode(nodeValue string) string {
	// This is a simplified encoder - in real usage, use node.NewEncoder
	varReg := variable.VarReg
	idx := 0
	seen := make(map[string]string)

	return varReg.ReplaceAllStringFunc(nodeValue, func(rawVar string) string {
		if alias, ok := seen[rawVar]; ok {
			return alias
		}
		alias := fmt.Sprintf("[VAR-%d]", idx)
		seen[rawVar] = alias
		idx++
		return alias
	})
}

func getProjectRoot() string {
	// Try to find project root from current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Walk up to find go.mod
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "."
		}
		wd = parent
	}
}

// testRepository implements container.Repository for testing
type testRepository struct {
	data map[string]m.Hash
}

func (r *testRepository) HGet(tag string) (m.Hash, error) {
	if val, ok := r.data[tag]; ok {
		return val, nil
	}
	return nil, nil
}

func (r *testRepository) HSet(tag string, value m.Hash) error {
	r.data[tag] = value
	return nil
}

// Ensure testRepository implements Repository
var _ Repository = (*testRepository)(nil)

// DeleteCycle removes a container and all its related entries from the storage
// This breaks the cycle by removing the container that participates in the circular relationship
func DeleteCycle(repo Repository, encoder node.Encoder, containerID string) error {
	// Get the container's triggers and results before deletion
	tTag := prefixC2T + containerID
	rTag := prefixC2R + containerID

	triggers, _ := repo.HGet(tTag)
	results, _ := repo.HGet(rTag)

	// Remove T->C mappings (trigger -> container)
	for trigger := range triggers {
		encodedTrigger := encoder.Do(trigger)
		t2cTag := prefixT2C + encodedTrigger
		if containers, err := repo.HGet(t2cTag); err == nil && containers != nil {
			delete(containers, containerID)
			if len(containers) == 0 {
				// If no more containers for this trigger, delete the entire entry
				repo.HSet(t2cTag, nil)
			} else {
				repo.HSet(t2cTag, containers)
			}
		}
	}

	// Remove R->C mappings (result -> container)
	for result := range results {
		encodedResult := encoder.Do(result)
		r2cTag := prefixR2C + encodedResult
		if containers, err := repo.HGet(r2cTag); err == nil && containers != nil {
			delete(containers, containerID)
			if len(containers) == 0 {
				repo.HSet(r2cTag, nil)
			} else {
				repo.HSet(r2cTag, containers)
			}
		}
	}

	// Remove C->T and C->R mappings
	repo.HSet(tTag, nil)
	repo.HSet(rTag, nil)

	return nil
}

// FindAndDeleteCycles detects all cycles and deletes the containers causing them
// Returns the list of deleted container IDs
func FindAndDeleteCycles(repo Repository, encoder node.Encoder) ([]string, error) {
	detector := NewCycleDetector(repo, encoder)

	// Find all cycles
	cycles, err := detector.DetectExistingCycles()
	if err != nil {
		return nil, fmt.Errorf("failed to detect cycles: %w", err)
	}

	if len(cycles) == 0 {
		return nil, nil
	}

	// Collect unique container IDs from all cycles
	containersToDelete := make(map[string]bool)
	for _, cycle := range cycles {
		for _, step := range cycle.Path {
			// Extract container ID from path steps like "container[abc123]"
			if strings.HasPrefix(step, "container[") && strings.HasSuffix(step, "]") {
				containerID := step[len("container[") : len(step)-1]
				containersToDelete[containerID] = true
			}
		}
	}

	// Delete each container involved in cycles
	var deleted []string
	for containerID := range containersToDelete {
		if err := DeleteCycle(repo, encoder, containerID); err != nil {
			return deleted, fmt.Errorf("failed to delete container %s: %w", containerID, err)
		}
		deleted = append(deleted, containerID)
	}

	return deleted, nil
}

// TestDeleteCycles finds and deletes all cycles from the data
func TestDeleteCycles(t *testing.T) {
	dataPath := filepath.Join(getProjectRoot(), "archive", "Data", "json", "hash_data.json")

	data, err := os.ReadFile(dataPath)
	if err != nil {
		t.Fatalf("Failed to read hash_data.json: %v", err)
	}

	var hData map[string]m.Hash
	if err := json.Unmarshal(data, &hData); err != nil {
		t.Fatalf("Failed to unmarshal hash_data.json: %v", err)
	}

	// Create a modifiable copy of the data
	repo := &testRepository{data: make(map[string]m.Hash)}
	for k, v := range hData {
		repo.data[k] = v
	}

	// Create encoder
	varReg := variable.VarReg
	encoder, err := node.NewEncoder(varReg)
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}

	// First, detect cycles
	detector := NewCycleDetector(repo, encoder)
	cycles, err := detector.DetectExistingCycles()
	if err != nil {
		t.Fatalf("Failed to detect cycles: %v", err)
	}

	if len(cycles) == 0 {
		t.Log("✓ No cycles found - nothing to delete")
		return
	}

	t.Logf("⚠ Found %d cycle(s) to delete:", len(cycles))
	for i, cycle := range cycles {
		t.Logf("\nCycle %d:", i+1)
		for _, step := range cycle.Path {
			t.Logf("  → %s", step)
		}
	}

	// Delete cycles
	deleted, err := FindAndDeleteCycles(repo, encoder)
	if err != nil {
		t.Fatalf("Failed to delete cycles: %v", err)
	}

	t.Logf("\n✓ Deleted %d container(s) involved in cycles:", len(deleted))
	for _, id := range deleted {
		t.Logf("  - %s", id)
	}

	// Verify no cycles remain
	cyclesAfter, err := detector.DetectExistingCycles()
	if err != nil {
		t.Fatalf("Failed to verify cycles after deletion: %v", err)
	}

	if len(cyclesAfter) > 0 {
		t.Errorf("⚠ Still found %d cycle(s) after deletion:", len(cyclesAfter))
		for i, cycle := range cyclesAfter {
			t.Logf("\nRemaining Cycle %d:", i+1)
			for _, step := range cycle.Path {
				t.Logf("  → %s", step)
			}
		}
	} else {
		t.Log("✓ All cycles successfully removed")
	}
}

// TestWithRealStorage tests cycle detection using the actual JSON storage
func TestWithRealStorage(t *testing.T) {
	dataPath := filepath.Join(getProjectRoot(), "archive", "Data", "json", "hash_data.json")

	// Create real JSON repository
	repo := storage.NewJSONRepo(dataPath, "")

	// Initialize with data
	data, err := os.ReadFile(dataPath)
	if err != nil {
		t.Fatalf("Failed to read hash_data.json: %v", err)
	}

	var hData map[string]m.Hash
	if err := json.Unmarshal(data, &hData); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Populate the repo
	for k, v := range hData {
		repo.HSet(k, v)
	}

	// Create encoder
	varReg := variable.VarReg
	encoder, err := node.NewEncoder(varReg)
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}

	// Create detector
	detector := NewCycleDetector(repo, encoder)

	// Detect cycles
	cycles, err := detector.DetectExistingCycles()
	if err != nil {
		t.Fatalf("Cycle detection failed: %v", err)
	}

	if len(cycles) == 0 {
		t.Log("✓ No cycles detected in hash_data.json")
	} else {
		t.Errorf("⚠ Found %d cycle(s):", len(cycles))
		for i, cycle := range cycles {
			t.Logf("\nCycle %d path:", i+1)
			for j, step := range cycle.Path {
				t.Logf("  %d. %s", j+1, step)
			}
		}
	}
}
