package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// FindAllJsonFiles find all "*.json" files under a given dir
func FindAllJsonFiles(dir string) []string {
	var jsonFiles []string

	walkErr := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.ToLower(filepath.Ext(d.Name())) == ".json" {
			jsonFiles = append(jsonFiles, path)
		}

		return nil
	})

	if walkErr != nil {
		return nil
	}

	return jsonFiles
}
