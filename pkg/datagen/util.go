package datagen

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type JsonFormatData struct {
	Triggers []string `json:"triggers"`
	Results  []string `json:"results"`
	Tag      string   `json:"tag"`
}

func saveToJsonWithNamespace(item Data, path string, namespace string) error {
	jsonStructured := &JsonFormatData{
		Triggers: item.T(),
		Results:  item.R(),
		Tag:      namespace,
	}
	if path == "" {
		return errors.New("path is required")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	jsonData, err := json.MarshalIndent(jsonStructured, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON marshal failed: %w", err)
	}

	tempPath := path + ".tmp"

	if err := os.WriteFile(tempPath, jsonData, 0644); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("write temp failed: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("rename failed: %w", err)
	}

	return nil
}
func makePath(namespace, filename string) string {
	if strings.TrimSpace(namespace) == "" {
		return ""
	}
	if strings.TrimSpace(filename) == "" {
		return ""
	}

	filename = strings.TrimSpace(filename)
	if !strings.HasSuffix(strings.ToLower(filename), jsonExt) {
		filename += jsonExt
	}

	fullPath := filepath.Join(
		RootDir,
		namespace,
		filename,
	)
	fullPath = filepath.Clean(fullPath)

	return fullPath
}
