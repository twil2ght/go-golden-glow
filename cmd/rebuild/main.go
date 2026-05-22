package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"goldenglow/m"
	"goldenglow/pkg/brainsaver"
	"goldenglow/pkg/database"
	"goldenglow/utils"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

var (
	mappingPath  = filepath.Join(utils.RootDir, "config/builder_mapping.json")
	trackingPath = filepath.Join(utils.RootDir, "archive/Data/json/container_sources.json")
)

type dataSection struct {
	Commands []string `json:"commands"`
}

type srcFile struct {
	Dependencies []string      `json:"dependencies"`
	Data         []dataSection `json:"data"`
	IsTemplate   bool          `json:"is_template"`
}

type containerData struct {
	T    m.Hash
	R    m.Hash
	Hash string
}

type sourceMap map[string][]string

func main() {
	dirs := os.Args[1:]
	if len(dirs) == 0 {
		dirs = []string{filepath.Join(utils.RootDir, "src")}
	}

	mapping := loadMapping()
	fmt.Println("mapping loaded:", mapping)

	repo := database.DefaultJSONRepo()
	if err := repo.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "init repo: %v\n", err)
		os.Exit(1)
	}
	saver := brainsaver.NewService(repo)

	_, statErr := os.Stat(trackingPath)
	bootstrap := statErr != nil
	tracking := loadTracking()

	if bootstrap {
		fmt.Println("(first run: populating tracking file, no database changes)")
	}

	var files []string
	for _, dir := range dirs {
		fullPath := dir
		if !filepath.IsAbs(dir) {
			fullPath = filepath.Join(utils.RootDir, dir)
		}
		files = append(files, utils.FindAllJsonFiles(fullPath)...)
	}

	totalAdded, totalRemoved := 0, 0

	for _, file := range files {
		containers := processFile(file, mapping)
		if containers == nil {
			continue
		}

		relPath, _ := filepath.Rel(utils.RootDir, file)
		relPath = filepath.ToSlash(relPath)

		var newHashes []string
		for _, c := range containers {
			newHashes = append(newHashes, c.Hash)
		}
		oldHashes := tracking[relPath]

		// Remove stale containers
		for _, old := range oldHashes {
			if !slices.Contains(newHashes, old) {
				removeContainer(repo, old)
				fmt.Printf("  removed: %s\n", old)
				totalRemoved++
			}
		}

		// Save new containers
		for _, c := range containers {
			if !slices.Contains(oldHashes, c.Hash) {
				if !bootstrap {
					saver.Save(c.T, c.R)
				}
				fmt.Printf("  added:   %s\n", c.Hash)
				totalAdded++
			}
		}

		if len(newHashes) > 0 || len(oldHashes) > 0 {
			fmt.Printf("%s: %d containers\n", relPath, len(newHashes))
		}
		tracking[relPath] = newHashes
	}

	// Clean up tracking entries for deleted files
	for path := range tracking {
		fullPath := filepath.Join(utils.RootDir, filepath.FromSlash(path))
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			fmt.Printf("file gone, removing containers: %s\n", path)
			for _, hash := range tracking[path] {
				removeContainer(repo, hash)
				totalRemoved++
			}
			delete(tracking, path)
		}
	}

	saveTracking(tracking)
	if err := repo.Shutdown(); err != nil {
		fmt.Fprintf(os.Stderr, "save repo: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nDone. Added: %d, Removed: %d\n", totalAdded, totalRemoved)
}

func processFile(file string, mapping map[string]string) []containerData {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil
	}

	var sf srcFile
	if err := json.Unmarshal(content, &sf); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: parse %s: %v\n", file, err)
		return nil
	}

	if sf.IsTemplate {
		return nil
	}

	var results []containerData

	for _, section := range sf.Data {
		if len(section.Commands) == 0 {
			continue
		}
		containers := simulateBuilder(section.Commands, mapping)
		results = append(results, containers...)
	}

	return results
}

func simulateBuilder(commands []string, mapping map[string]string) []containerData {
	var (
		results     []containerData
		input       []string
		inputSingle []string
		buildDone   bool
	)

	for _, cmd := range commands {
		var valueType, value string

		switch {
		case strings.HasPrefix(cmd, "[input] "):
			value = strings.TrimPrefix(cmd, "[input] ")
			valueType = "input"
		case strings.HasPrefix(cmd, "[output] "):
			value = strings.TrimPrefix(cmd, "[output] ")
			valueType = "output"
		default:
			continue
		}

		value = mapToPlaceholder(value, mapping)

		if valueType == "output" {
			buildDone = true
			if value != "[clear]" {
				if len(input) > 0 {
					results = append(results, containerData{
						T: m.ToHash(copySlice(input)),
						R: m.ToHash([]string{value}),
					})
				}
				if len(input) == 1 && len(inputSingle) > 0 {
					for _, s := range inputSingle {
						results = append(results, containerData{
							T: m.ToHash([]string{s}),
							R: m.ToHash([]string{value}),
						})
					}
				}
			}
			continue
		}

		if buildDone {
			input = nil
			inputSingle = nil
			buildDone = false
		}

		input = append(input, value)
	}

	// Compute hashes
	for i := range results {
		results[i].Hash = genHashForTR(results[i].T, results[i].R)
	}

	return results
}

func mapToPlaceholder(value string, mapping map[string]string) string {
	parts := strings.Fields(value)
	for i, part := range parts {
		if placeholder, exists := mapping[part]; exists {
			parts[i] = placeholder
		}
	}
	return strings.Join(parts, " ")
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

func removeContainer(repo database.Repository, hashKey string) {
	const (
		prefixC2T = "C->T:"
		prefixC2R = "C->R:"
		prefixT2C = "T->C:"
		prefixR2C = "R->C:"
	)

	tNodes, _ := repo.HGet(prefixC2T + hashKey)
	rNodes, _ := repo.HGet(prefixC2R + hashKey)

	repo.HDel(prefixC2T + hashKey)
	repo.HDel(prefixC2R + hashKey)

	for t := range tNodes {
		repo.HDel(prefixT2C+t, hashKey)
	}
	for r := range rNodes {
		repo.HDel(prefixR2C+r, hashKey)
	}
}

func loadMapping() map[string]string {
	mapping := make(map[string]string)
	data, err := os.ReadFile(mappingPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: read mapping: %v\n", err)
		return mapping
	}
	if err := json.Unmarshal(data, &mapping); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: parse mapping: %v\n", err)
	}
	return mapping
}

func loadTracking() sourceMap {
	tracking := make(sourceMap)
	data, err := os.ReadFile(trackingPath)
	if err != nil {
		return tracking
	}
	if err := json.Unmarshal(data, &tracking); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: parse tracking: %v\n", err)
	}
	return tracking
}

func saveTracking(tracking sourceMap) {
	// Remove empty entries
	for path, hashes := range tracking {
		if len(hashes) == 0 {
			delete(tracking, path)
		}
	}
	data, err := json.MarshalIndent(tracking, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: marshal tracking: %v\n", err)
		return
	}
	if err := os.WriteFile(trackingPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: write tracking: %v\n", err)
	}
}

func copySlice(s []string) []string {
	out := make([]string, len(s))
	copy(out, s)
	return out
}
