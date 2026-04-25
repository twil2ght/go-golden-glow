package main

import (
	"fmt"
	"os"
	"strings"

	"goldenglow/pkg/container/fetcher"
	"goldenglow/pkg/database"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run cmd/tool/main.go <node-value>")
		os.Exit(1)
	}
	nodeValue := strings.Join(os.Args[1:], " ")

	repo := database.DefaultJSONRepo()
	if err := repo.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to init database: %v\n", err)
		os.Exit(1)
	}
	defer repo.Shutdown()

	f := fetcher.Default()
	if f == nil {
		fmt.Fprintln(os.Stderr, "failed to create fetcher")
		os.Exit(1)
	}

	containerHashes := make(map[string]struct{})

	// Look up as T-node
	if t2c, err := repo.HGet("T->C:" + nodeValue); err == nil {
		for h := range t2c {
			containerHashes[h] = struct{}{}
		}
	}

	// Look up as R-node
	if r2c, err := repo.HGet("R->C:" + nodeValue); err == nil {
		for h := range r2c {
			containerHashes[h] = struct{}{}
		}
	}

	if len(containerHashes) == 0 {
		fmt.Printf("node %q not found in any container\n", nodeValue)
		return
	}

	fmt.Printf("node %q found in %d container(s):\n\n", nodeValue, len(containerHashes))

	i := 1
	for hash := range containerHashes {
		fmt.Printf("── Container %d (hash: %s)\n", i, hash)
		i++

		tNodes := f.T(hash)
		rNodes := f.R(hash)

		fmt.Println("   T (triggers):")
		if len(tNodes) == 0 {
			fmt.Println("      (none)")
		}
		for _, n := range tNodes {
			marker := ""
			if n.Value() == nodeValue {
				marker = " ◀◀◀"
			}
			fmt.Printf("      • %s%s\n", n.Value(), marker)
		}

		fmt.Println("   R (results):")
		if len(rNodes) == 0 {
			fmt.Println("      (none)")
		}
		for _, n := range rNodes {
			marker := ""
			if n.Value() == nodeValue {
				marker = " ◀◀◀"
			}
			fmt.Printf("      • %s%s\n", n.Value(), marker)
		}
		fmt.Println()
	}
}
