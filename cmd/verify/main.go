package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"goldenglow/m"
	"goldenglow/pkg/database"
	"goldenglow/pkg/datagen"
	"goldenglow/utils"
)

// builderMapping mirrors plugin/builtin/builder/mapping.json
var builderMapping = map[string]string{
	"A":  "$1",
	"B":  "$2",
	"C":  "$3",
	"D":  "$4",
	"E":  "$5",
	"F":  "$6",
	"A1": "$7",
	"E1": "$8",
}

const (
	prefixInput  = "[input] "
	prefixOutput = "[output] "
)

func main() {
	repo := database.DefaultJSONRepo()
	if err := repo.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to init database: %v\n", err)
		os.Exit(1)
	}
	defer repo.Shutdown()

	exitCode := 0

	// Always verify generated data containers
	if !verifyGenerated(repo) {
		exitCode = 1
	}

	// Verify builder containers from archive directories passed as args
	for _, dir := range os.Args[1:] {
		fmt.Printf("\nArchive: %s\n", dir)
		if !verifyArchive(repo, dir) {
			exitCode = 1
		}
	}

	if exitCode != 0 {
		os.Exit(1)
	}
}

func verifyGenerated(repo database.Repository) bool {
	jsonFiles := utils.FindAllJsonFiles("generated")
	if len(jsonFiles) == 0 {
		fmt.Println("no generated files found")
		return false
	}

	var total, found, missing int
	var failures []string

	for _, jsonFile := range jsonFiles {
		content, err := os.ReadFile(jsonFile)
		if err != nil {
			continue
		}

		var data datagen.JsonFormatData
		if err := json.Unmarshal(content, &data); err != nil {
			continue
		}

		if len(data.Triggers) == 0 && len(data.Results) == 0 {
			continue
		}

		total++
		tHash := m.ToHash(data.Triggers)
		rHash := m.ToHash(data.Results)
		expectedHash := genHashForTR(tHash, rHash)

		if !checkContainer(repo, expectedHash, tHash, rHash) {
			missing++
			failures = append(failures, fmt.Sprintf("  MISSING: %s (hash: %s)", jsonFile, expectedHash))
			continue
		}
		found++
	}

	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("  generated files checked: %d\n", total)
	fmt.Printf("  containers found:        %d\n", found)
	if missing > 0 {
		fmt.Printf("  containers missing:      %d\n", missing)
		fmt.Println(strings.Repeat("─", 60))
		for _, f := range failures {
			fmt.Println(f)
		}
		fmt.Println("\nVERDICT: FAIL")
		return false
	}
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println("VERDICT: PASS — all generated containers saved")
	return true
}

func verifyArchive(repo database.Repository, dir string) bool {
	jsonFiles := utils.FindAllJsonFiles(dir)
	if len(jsonFiles) == 0 {
		fmt.Printf("  no JSON files found in %q\n", dir)
		return false
	}

	var total, found, missing int
	var failures []string

	for _, jsonFile := range jsonFiles {
		content, err := os.ReadFile(jsonFile)
		if err != nil {
			continue
		}

		expectedContainers := parseCommands(content)
		for _, ec := range expectedContainers {
			total++
			tHash := m.ToHash(ec.t)
			rHash := m.ToHash(ec.r)
			expectedHash := genHashForTR(tHash, rHash)

			if !checkContainer(repo, expectedHash, tHash, rHash) {
				missing++
				failures = append(failures, fmt.Sprintf(
					"  MISSING: %s [%s]\n    T: %v\n    R: %v\n    hash: %s",
					jsonFile, ec.desc, ec.t, ec.r, expectedHash,
				))
			} else {
				found++
			}
		}
	}

	if total == 0 {
		fmt.Println("  no builder [input]/[output] pairs found")
		return true
	}

	fmt.Printf("  builder containers expected: %d\n", total)
	fmt.Printf("  containers found:            %d\n", found)
	if missing > 0 {
		fmt.Printf("  containers missing:          %d\n", missing)
		fmt.Println()
		for _, f := range failures {
			fmt.Println(f)
		}
		fmt.Println("\nVERDICT: FAIL")
		return false
	}
	fmt.Println("VERDICT: PASS — all builder containers saved")
	return true
}

type expectedContainer struct {
	t    []string
	r    []string
	desc string
}

func parseCommands(content []byte) []expectedContainer {
	var result []expectedContainer

	// Try array of objects first
	var items []struct {
		Description string   `json:"description"`
		Commands    []string `json:"commands"`
	}
	if json.Unmarshal(content, &items) == nil && len(items) > 0 {
		for _, item := range items {
			result = append(result, extractContainers(item.Commands, item.Description)...)
		}
		return result
	}

	// Fall back to single object
	var single struct {
		Description string   `json:"description"`
		Commands    []string `json:"commands"`
	}
	if json.Unmarshal(content, &single) == nil && len(single.Commands) > 0 {
		result = append(result, extractContainers(single.Commands, single.Description)...)
	}

	return result
}

func extractContainers(commands []string, desc string) []expectedContainer {
	var result []expectedContainer
	var inputs []string

	for _, cmd := range commands {
		if after, ok := strings.CutPrefix(cmd, prefixInput); ok {
			inputs = append(inputs, mapValue(after))
			continue
		}
		if after, ok := strings.CutPrefix(cmd, prefixOutput); ok {
			if len(inputs) > 0 {
				result = append(result, expectedContainer{
					t:    copySlice(inputs),
					r:    []string{mapValue(after)},
					desc: desc,
				})
			}
			inputs = nil
		}
	}
	return result
}

func mapValue(s string) string {
	parts := strings.Fields(s)
	for i, p := range parts {
		if placeholder, ok := builderMapping[p]; ok {
			parts[i] = placeholder
		}
	}
	return strings.Join(parts, " ")
}

func checkContainer(repo database.Repository, hash string, expectedT, expectedR m.Hash) bool {
	actualT, err := repo.HGet("C->T:" + hash)
	if err != nil || actualT == nil {
		return false
	}
	if !hashEqual(expectedT, actualT) {
		return false
	}
	actualR, err := repo.HGet("C->R:" + hash)
	if err != nil || actualR == nil {
		return false
	}
	return hashEqual(expectedR, actualR)
}

func genHashForTR(tv, rv m.Hash) string {
	sortedT := sortedKeys(tv)
	sortedR := sortedKeys(rv)
	str := fmt.Sprintf("T:[%s]|R:[%s]", sortedT, sortedR)
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

func sortedKeys(m m.Hash) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

func hashEqual(a, b m.Hash) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

func copySlice(s []string) []string {
	out := make([]string, len(s))
	copy(out, s)
	return out
}
