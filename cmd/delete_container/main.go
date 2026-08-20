package main

import (
	"fmt"
	"goldenglow/pkg/database"
	"os"
)

var (
	prefixC2T = "C->T:"
	prefixC2R = "C->R:"
	prefixT2C = "T->C:"
	prefixR2C = "R->C:"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run cmd/delete_container/main.go <hash>")
		os.Exit(1)
	}
	hash := os.Args[1]

	repo := database.DefaultJSONRepo()
	if err := repo.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to init database: %v\n", err)
		os.Exit(1)
	}
	defer repo.Shutdown()

	tNodes, err := repo.HGet(prefixC2T + hash)
	if err != nil {
		tNodes = nil
	}
	rNodes, err := repo.HGet(prefixC2R + hash)
	if err != nil {
		rNodes = nil
	}

	if len(tNodes) == 0 && len(rNodes) == 0 {
		fmt.Printf("container %q not found\n", hash)
		return
	}

	// Remove this container from each T-node's reverse index
	for tVal := range tNodes {
		repo.HDel(prefixT2C+tVal, hash)
	}

	// Remove this container from each R-node's reverse index
	for rVal := range rNodes {
		repo.HDel(prefixR2C+rVal, hash)
	}

	// Delete the container's T and R entries
	repo.HDel(prefixC2T + hash)
	repo.HDel(prefixC2R + hash)

	fmt.Printf("deleted container %q (%d T-nodes, %d R-nodes)\n", hash, len(tNodes), len(rNodes))
}
