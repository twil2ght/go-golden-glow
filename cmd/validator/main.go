package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var headPattern = regexp.MustCompile(`\[[a-zA-Z][a-zA-Z0-9_]*\]`)

// excludedHeads are command-type prefixes, not plugin handler heads.
// They don't satisfy the "-> requires a head" rule on their own.
var excludedHeads = map[string]bool{
	"[input]":  true,
	"[output]": true,
}

func main() {
	dir := "archive/logic"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	var violations []string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			violations = append(violations, fmt.Sprintf("%s: error walking: %v", path, err))
			return nil
		}
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		checkFile(path, &violations)
		return nil
	})

	if len(violations) == 0 {
		fmt.Println("All data files pass validation.")
		os.Exit(0)
	}

	fmt.Printf("Found %d violation(s):\n\n", len(violations))
	for _, v := range violations {
		fmt.Println("  " + v)
	}
	fmt.Println()
	os.Exit(1)
}

// checkFile reads a JSON file and validates all command strings within it.
// Handles three JSON shapes: array of objects, single object, or array of strings.
func checkFile(path string, violations *[]string) {
	data, err := os.ReadFile(path)
	if err != nil {
		*violations = append(*violations, fmt.Sprintf("%s: cannot read: %v", path, err))
		return
	}

	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return
	}

	switch trimmed[0] {
	case '[':
		var entries []json.RawMessage
		if err := json.Unmarshal(data, &entries); err != nil {
			*violations = append(*violations, fmt.Sprintf("%s: invalid JSON: %v", path, err))
			return
		}
		for i, entry := range entries {
			checkEntry(path, i, entry, violations)
		}
	case '{':
		checkEntry(path, -1, data, violations)
	default:
		*violations = append(*violations, fmt.Sprintf("%s: unexpected JSON root type", path))
	}
}

// checkEntry validates an individual JSON object (or raw value) containing commands.
func checkEntry(path string, idx int, raw json.RawMessage, violations *[]string) {
	prefix := ""
	if idx >= 0 {
		prefix = fmt.Sprintf("entry[%d] ", idx)
	}

	var obj struct {
		Commands []string `json:"commands"`
		In       string   `json:"in"`
		Out      any      `json:"out"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		// Maybe a plain string array (test files etc.)
		var strs []string
		if err2 := json.Unmarshal(raw, &strs); err2 == nil {
			for _, cmd := range strs {
				checkCommand(path, prefix, cmd, violations)
			}
			return
		}
		// Maybe a plain string
		var str string
		if err3 := json.Unmarshal(raw, &str); err3 == nil {
			checkCommand(path, prefix, str, violations)
			return
		}
		*violations = append(*violations, fmt.Sprintf("%s: %scannot parse entry: %v", path, prefix, err))
		return
	}

	for _, cmd := range obj.Commands {
		checkCommand(path, prefix, cmd, violations)
	}
	checkCommand(path, prefix, obj.In, violations)
	switch v := obj.Out.(type) {
	case string:
		checkCommand(path, prefix, v, violations)
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok {
				checkCommand(path, prefix, s, violations)
			}
		}
	}

	// Check src/test prefix conventions
	isSrc := strings.Contains(filepath.ToSlash(path), "/src/")
	isTest := strings.Contains(filepath.ToSlash(path), "/test/")
	for _, cmd := range obj.Commands {
		if isSrc && !strings.HasPrefix(cmd, "[input]") && !strings.HasPrefix(cmd, "[output]") {
			*violations = append(*violations, fmt.Sprintf("%s: %scommand in src/ does not start with [input] or [output]: %q", path, prefix, cmd))
		}
		if isTest && (strings.HasPrefix(cmd, "[input]") || strings.HasPrefix(cmd, "[output]")) {
			*violations = append(*violations, fmt.Sprintf("%s: %scommand in test/ should not use [input]/[output] tag: %q", path, prefix, cmd))
		}
	}
}

// checkCommand applies validation rules to a single command string.
func checkCommand(path, prefix, cmd string, violations *[]string) {
	if cmd == "" {
		return
	}
	if !strings.Contains(cmd, "->") {
		return
	}
	// Must have at least one non-excluded head (e.g. [repo], [time], [ST], custom heads)
	matches := headPattern.FindAllString(cmd, -1)
	for _, m := range matches {
		if !excludedHeads[m] {
			return // found a valid handler head
		}
	}
	*violations = append(*violations, fmt.Sprintf("%s: %s-> without handler [head]: %q", path, prefix, cmd))
}
