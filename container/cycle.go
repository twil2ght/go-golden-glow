package container

import (
	"fmt"
	"goldenglow/m"
	"goldenglow/node"
)

// CycleDetector detects cycles in container trigger-result relationships
// A cycle occurs when:
// - Container X has Trigger A with Result B
// - Container Y has Trigger B with Result A (or any chain that loops back)
type CycleDetector struct {
	db      Repository
	encoder node.Encoder
}

// CycleInfo holds information about a detected cycle
type CycleInfo struct {
	Path []string // The cycle path: container1 -> node1 -> container2 -> node2 -> ...
}

func (c *CycleInfo) Error() string {
	return fmt.Sprintf("cycle detected: %v", c.Path)
}

// NewCycleDetector creates a new cycle detector
func NewCycleDetector(db Repository, encoder node.Encoder) *CycleDetector {
	return &CycleDetector{
		db:      db,
		encoder: encoder,
	}
}

// DetectCycle checks if adding a container with the given triggers and results would create a cycle
// tv: trigger nodes (TNodes) - these are inputs that trigger the container
// rv: result nodes (RNodes) - these are outputs produced by the container
func (cd *CycleDetector) DetectCycle(tv, rv m.Hash) error {
	// For each result node, check if it can trigger a chain that leads back to any trigger node
	for rNode := range rv {
		visited := make(map[string]bool)
		path := []string{fmt.Sprintf("result[%s]", rNode)}
		if err := cd.detectFromNode(rNode, visited, path, tv); err != nil {
			return err
		}
	}
	return nil
}

// detectFromNode performs DFS from a result node to find cycles
// It checks if this result node (as a trigger) leads to containers that eventually produce
// any of the original trigger nodes
func (cd *CycleDetector) detectFromNode(nodeValue string, visited map[string]bool, path []string, targetTriggers m.Hash) error {
	encodedNode := cd.encoder.Do(nodeValue)

	// Check if this node is in the target triggers (cycle found)
	if _, ok := targetTriggers[nodeValue]; ok && len(path) > 1 {
		return &CycleInfo{Path: append(path, fmt.Sprintf("trigger[%s]", nodeValue))}
	}

	// Prevent infinite recursion
	if visited[encodedNode] {
		return nil
	}
	visited[encodedNode] = true

	// Find all containers that have this node as a trigger (T->C mapping)
	// This means: the current node triggers these containers
	tag := prefixT2C + encodedNode
	containerMap, err := cd.db.HGet(tag)
	if err != nil {
		return nil // No containers found for this node
	}

	// For each container that this node triggers, check their results
	for containerID := range containerMap {
		if err := cd.detectFromContainer(containerID, visited, append(path, fmt.Sprintf("container[%s]", containerID)), targetTriggers); err != nil {
			return err
		}
	}

	return nil
}

// detectFromContainer checks all result nodes of a container
func (cd *CycleDetector) detectFromContainer(containerID string, visited map[string]bool, path []string, targetTriggers m.Hash) error {
	// Get all result nodes of this container (C->R mapping)
	rTag := prefixC2R + containerID
	rNodes, err := cd.db.HGet(rTag)
	if err != nil {
		return nil
	}

	// For each result node, continue DFS
	for rNode := range rNodes {
		if err := cd.detectFromNode(rNode, visited, append(path, fmt.Sprintf("result[%s]", rNode)), targetTriggers); err != nil {
			return err
		}
	}

	return nil
}

// DetectExistingCycles checks the entire storage for any existing cycles
func (cd *CycleDetector) DetectExistingCycles() ([]CycleInfo, error) {
	// Get all containers
	nodeSet, err := cd.db.HGet(KeyNodeSet)
	if err != nil {
		return nil, err
	}

	var cycles []CycleInfo
	checkedContainers := make(map[string]bool)

	// For each node, find containers it triggers and check for cycles
	for nodeValue := range nodeSet {
		encodedNode := cd.encoder.Do(nodeValue)
		tag := prefixT2C + encodedNode
		containerMap, err := cd.db.HGet(tag)
		if err != nil {
			continue
		}

		for containerID := range containerMap {
			if checkedContainers[containerID] {
				continue
			}
			checkedContainers[containerID] = true

			// Get triggers and results for this container
			tTag := prefixC2T + containerID
			rTag := prefixC2R + containerID

			triggers, _ := cd.db.HGet(tTag)
			results, _ := cd.db.HGet(rTag)

			if triggers == nil || results == nil {
				continue
			}

			if err := cd.DetectCycle(triggers, results); err != nil {
				if cycleInfo, ok := err.(*CycleInfo); ok {
					cycles = append(cycles, *cycleInfo)
				}
			}
		}
	}

	return cycles, nil
}
