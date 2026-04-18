package dataloader

import (
	"goldenglow/m"
	"os"
	"path/filepath"
	"strings"
)

// findAllJsonFiles find all "*.json" files under a given dir
func findAllJsonFiles(dir string) []string {
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
func OnLoadingError(file string) {
	logger.Error("loading json data err: ",
		"module", "data loader",
		"file", file,
	)
}
func OnFinishedInfoSum(tagMap m.Map[int]) {
	logger.Info(" data loading is Done",
		"module", "data loader",
		"plugin amount", len(tagMap),
	)
	for tag, count := range tagMap {
		logger.Info("info:",
			"module", "data loader",
			"plugin tag", tag,
			"file count", count,
		)
	}
}
