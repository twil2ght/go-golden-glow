package builder

import (
	"encoding/json"
	"errors"
	"goldenglow/m"
	"os"
	"path/filepath"
	"strings"

	"goldenglow/utils"
)

var (
	templateDir = filepath.Join(utils.RootDir, "src", "Grammar")
)

type template struct {
	Name       string     `json:"name"`
	IsTemplate bool       `json:"is_template"`
	Args       []string   `json:"args"`
	Data       []dataItem `json:"data"`
}

type dataItem struct {
	Commands []string `json:"commands"`
}

func loadTemplates(templateDir string) (m.Map[*template], error) {
	templates := make(map[string]*template)
	_ = filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var tpl *template
		if err := json.Unmarshal(data, &tpl); err != nil {
			logger.Warn("templateGen: failed to parse", "file", path, "err", err)
		}
		if tpl.IsTemplate && tpl.Name != "" {
			templates[tpl.Name] = tpl
			logger.Debug("templateGen: loaded template", "name", tpl.Name, "file", path)
		}
		return nil
	})
	return templates, nil
}

func replaceVars(s string, vars map[string]string) string {
	for k, v := range vars {
		s = strings.ReplaceAll(s, k, v)
	}
	return s
}

// parseTemplateArgs args:"key1 = value1,key2 = value2" -> map[key1:value1 key2:value2]
func parseTemplateArgs(args string, tpl *template) (map[string]string, error) {
	result := make(map[string]string)
	if len(strings.Split(args, ",")) != len(tpl.Args) {
		logger.Debug("templateGen: wrong number of args", "expected", len(tpl.Args), "got", len(args))
		return nil, errors.New("templateGen: wrong number of args")
	}
	for i, value := range strings.Split(args, ",") {
		if value == "" {
			continue
		}
		result[strings.TrimSpace(tpl.Args[i])] = strings.TrimSpace(value)
		logger.Debug("templateGen: replaced arg", "key", tpl.Args[i], "arg", value)
	}
	return result, nil
}
