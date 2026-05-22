// Command cycledetector analyzes the container graph in archive/Data/json/hash_data.json
// to find cycles that could cause infinite loops at runtime.
//
// Usage:
//
//	go run ./cmd/cycledetector/
//
// It builds a directed graph where edge A→B means container A's result (R) nodes
// could match container B's trigger (T) templates. It then finds strongly connected
// components (cycles) and filters out known-safe iterator patterns (those involving
// [compute], [safe++], [CondGroup:Count++], etc.).
//
// Remaining cycles are reported as potential infinite-loop risks.
package main

import (
	"encoding/json"
	"fmt"
	"goldenglow/pkg/node/template"
	"goldenglow/utils"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var varRe = regexp.MustCompile(`\$\d+`)

// iteratorPatterns are keywords in T/R node values that indicate a cycle
// is part of the intentional iterator/step-advance design, not a runaway loop.
var iteratorPatterns = []string{
	"[compute]",
	"[safe++]",
	"[safe--]",
	"[CondGroup:Count++]",
	"[CondGroup:ElementCount++]",
	"[CondGroup:MarkDone",
	"[CondGroup:DoneAuto",
	"[CondGroup:Reset",
	"step $",
}

// nodeCategories classifies a node value for reporting
const (
	catExecutor  = "executor"
	catExtractor = "extractor"
	catChecker   = "checker"
	catPlain     = "plain"
)

func nodeCategory(val string) string {
	if strings.HasPrefix(val, "[node:executor]") {
		return catExecutor
	}
	if strings.HasPrefix(val, "[node:extractor]") {
		return catExtractor
	}
	if strings.HasPrefix(val, "[node:checker]") {
		return catChecker
	}
	return catPlain
}

// couldMatch determines whether a resolved R-node value could match a T-node template.
// It resolves R's $n variables to a placeholder ("X") and delegates to the engine's
// MatchTemplate. This may produce false positives (reporting matches that wouldn't
// occur at runtime) but should not produce false negatives.
func couldMatch(rValue, tValue string) bool {
	resolved := varRe.ReplaceAllString(rValue, "X")
	ok, _ := template.MatchTemplate(resolved, tValue)
	return ok
}

// hasIteratorPattern returns true if any node value in the given set contains
// a keyword associated with the intentional iterator design.
func hasIteratorPattern(nodes []string) bool {
	for _, n := range nodes {
		for _, pat := range iteratorPatterns {
			if strings.Contains(n, pat) {
				return true
			}
		}
	}
	return false
}

func main() {
	dataPath := filepath.Join(utils.RootDir, "archive/Data/json/hash_data.json")

	raw, err := os.ReadFile(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", dataPath, err)
		os.Exit(1)
	}

	var entries map[string]map[string]any
	if err := json.Unmarshal(raw, &entries); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", dataPath, err)
		os.Exit(1)
	}

	// Extract container T and R node sets.
	tNodes := map[string][]string{} // container hash → T node values
	rNodes := map[string][]string{} // container hash → R node values

	for key, vals := range entries {
		if strings.HasPrefix(key, "C->T:") {
			h := key[5:]
			for n := range vals {
				tNodes[h] = append(tNodes[h], n)
			}
		}
		if strings.HasPrefix(key, "C->R:") {
			h := key[5:]
			for n := range vals {
				rNodes[h] = append(rNodes[h], n)
			}
		}
	}

	// Only consider containers that have both T and R.
	hashes := make([]string, 0, len(tNodes))
	for h := range tNodes {
		if _, ok := rNodes[h]; ok {
			hashes = append(hashes, h)
		}
	}
	sort.Strings(hashes)
	fmt.Printf("Loaded %d containers\n\n", len(hashes))

	// Build adjacency: edge a → b exists when any R of a matches any T of b.
	adj := make(map[string]map[string]bool, len(hashes))
	for _, a := range hashes {
		adj[a] = make(map[string]bool)
		for _, r := range rNodes[a] {
			for _, b := range hashes {
				if adj[a][b] {
					continue // already have edge
				}
				for _, t := range tNodes[b] {
					if couldMatch(r, t) {
						adj[a][b] = true
						break
					}
				}
			}
		}
	}

	// Tarjan's algorithm to find strongly connected components.
	type scc struct {
		nodes []string
	}
	var sccs []scc

	stack := make([]string, 0, len(hashes))
	onStack := make(map[string]bool, len(hashes))
	dfn := make(map[string]int, len(hashes)) // discovery number
	low := make(map[string]int, len(hashes)) // lowlink
	counter := 0

	var strongconnect func(v string)
	strongconnect = func(v string) {
		counter++
		dfn[v] = counter
		low[v] = counter
		stack = append(stack, v)
		onStack[v] = true

		for w := range adj[v] {
			if dfn[w] == 0 {
				strongconnect(w)
				if low[w] < low[v] {
					low[v] = low[w]
				}
			} else if onStack[w] {
				if dfn[w] < low[v] {
					low[v] = dfn[w]
				}
			}
		}

		if low[v] == dfn[v] {
			comp := make([]string, 0)
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				comp = append(comp, w)
				if w == v {
					break
				}
			}
			if len(comp) > 1 || (len(comp) == 1 && adj[comp[0]][comp[0]]) {
				sccs = append(sccs, scc{nodes: comp})
			}
		}
	}

	for _, h := range hashes {
		if dfn[h] == 0 {
			strongconnect(h)
		}
	}

	if len(sccs) == 0 {
		fmt.Println("OK — no cycles detected.")
		return
	}

	// Classify and report each SCC.
	var dangerous int
	for _, scc := range sccs {
		allNodes := make([]string, 0, len(scc.nodes)*2)
		for _, h := range scc.nodes {
			allNodes = append(allNodes, tNodes[h]...)
			allNodes = append(allNodes, rNodes[h]...)
		}

		cats := make(map[string]int)
		for _, n := range allNodes {
			cats[nodeCategory(n)]++
		}

		isIterator := hasIteratorPattern(allNodes)
		isPureExtractor := cats[catExtractor] > 0 && cats[catExecutor] == 0 && cats[catChecker] == 0

		switch {
		case isIterator:
			if len(scc.nodes) == 1 {
				fmt.Printf("OK        self-loop  %s  (iterator pattern)\n", scc.nodes[0][:12])
			} else {
				fmt.Printf("OK        cycle %dd  (iterator pattern)\n", len(scc.nodes))
			}

		case isPureExtractor:
			if len(scc.nodes) == 1 {
				fmt.Printf("OK        self-loop  %s  (extractor, self-terminating)\n", scc.nodes[0][:12])
			} else {
				fmt.Printf("OK        cycle %dd  (extractor, self-terminating)\n", len(scc.nodes))
			}

		default:
			dangerous++
			if len(scc.nodes) == 1 {
				h := scc.nodes[0]
				fmt.Printf("⚠️  CYCLE   self-loop  %s\n", h[:12])
				for _, t := range tNodes[h] {
					cat := nodeCategory(t)
					fmt.Printf("           T [%-9s] %s\n", cat, t)
				}
				for _, r := range rNodes[h] {
					cat := nodeCategory(r)
					fmt.Printf("           R [%-9s] %s\n", cat, r)
				}
			} else {
				fmt.Printf("⚠️  CYCLE   %d-container chain\n", len(scc.nodes))
				for _, h := range scc.nodes {
					fmt.Printf("         %s\n", h[:12])
					for _, t := range tNodes[h] {
						cat := nodeCategory(t)
						fmt.Printf("           T [%-9s] %s\n", cat, t)
					}
					for _, r := range rNodes[h] {
						cat := nodeCategory(r)
						fmt.Printf("           R [%-9s] %s\n", cat, r)
					}
				}
			}
		}
	}

	if dangerous > 0 {
		fmt.Println()
		fmt.Printf("Found %d dangerous cycle(s).\n", dangerous)
		fmt.Println("Run `go run ./cmd/validator/` for additional data validation.")
		os.Exit(1)
	}
}
