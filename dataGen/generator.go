package dataGen

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type genBase struct {
	pluginName string
	langItems  map[string]Item
}

func NewGenerator(pluginName string) Generator {
	return &genBase{
		pluginName: pluginName,
		langItems:  make(map[string]Item),
	}
}

func (l *genBase) Add(langName string, langItem Item) {
	if langItem == nil {
		return
	}
	if langName == "" {
		return
	}
	l.langItems[langName] = langItem
}

func (l *genBase) Run() error {
	for name, e := range l.langItems {
		var (
			result   = l.rvGen(e.Params())
			path, _  = l.makePath(e.LangType(), l.pluginName, name+".json")
			jsonData = &JsonLangData{
				Triggers: e.Triggers(),
				Results:  []string{result},
			}
		)
		_ = l.save(jsonData, path)
	}
	return nil
}
func (l *genBase) rvGen(params Parameters) string {
	var (
		res = KeyDefault
	)
	for k, v := range params {
		res = fmt.Sprintf("%s [%s:%s]", res, k, v)
	}
	return res
}
func (l *genBase) save(dataItem *JsonLangData, path string) error {
	if path == "" {
		return errors.New("path is required")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	jsonData, err := json.MarshalIndent(dataItem, "", "  ")
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

func (l *genBase) makePath(langType LangType, pluginName, filename string) (string, error) {
	if strings.TrimSpace(pluginName) == "" {
		return "", errors.New("plugin name is required and cannot be empty")
	}
	if strings.TrimSpace(string(langType)) == "" {
		return "", errors.New("lang type is required and cannot be empty")
	}
	if strings.TrimSpace(filename) == "" {
		return "", errors.New("filename is required and cannot be empty")
	}

	filename = strings.TrimSpace(filename)
	if !strings.HasSuffix(strings.ToLower(filename), jsonExt) {
		filename += jsonExt
	}

	fullPath := filepath.Join(
		RootDir,
		pluginName,
		string(langType),
		filename,
	)
	fullPath = filepath.Clean(fullPath)

	return fullPath, nil
}
