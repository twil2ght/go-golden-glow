package builder

import (
	"encoding/json"
	"fmt"
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
	Name       string            `json:"name"`
	IsTemplate bool              `json:"is_template"`
	Args       map[string]string `json:"args"`
	Data       []dataItem        `json:"data"`
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
			return err
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

// parseTemplateArgs args:"key1=value1,key2=value2" -> map[key1:value1 key2:value2]
func parseTemplateArgs(args string) (map[string]string, error) {
	result := make(map[string]string)
	for _, pair := range strings.Split(args, ",") {
		if pair == "" {
			continue
		}
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid args pair: %s", pair)
		}
		result[kv[0]] = kv[1]
	}
	return result, nil
}
