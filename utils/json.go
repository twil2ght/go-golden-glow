package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// FindAllJsonFiles find all "*.json" files under a given dir
func FindAllJsonFiles(dir string) []string {
	var jsonFiles []string

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.ToLower(filepath.Ext(file.Name())) == ".json" {
			fullPath := filepath.Join(dir, file.Name())
			jsonFiles = append(jsonFiles, fullPath)
		}
	}

	return jsonFiles
}
