package lang

import (
	"encoding/json"
	"errors"
	"fmt"
	"goldenglow/dataGen"
	"goldenglow/m"
	"goldenglow/pkg/log"
	"goldenglow/utils"
	"os"
	"path/filepath"
	"strings"
)

var (
	logger = log.Default()
)

type langRegistry struct {
	pluginNameSet []string
	repo          Repo
}

func (l *langRegistry) Register(name string) error {
	if name == "" {
		return errors.New("empty name")
	}
	l.pluginNameSet = append(l.pluginNameSet, name)
	return nil
}

// RunAll 从每个插件目录下读取所有json文件并解析保存
func (l *langRegistry) RunAll() error {
	root := filepath.Join(utils.RootDir, "data")
	logger.Info("starting to parse langFiles",
		"total_plugins", len(l.pluginNameSet),
	)
	for _, name := range l.pluginNameSet {
		path := filepath.Join(root, name)

		jsonFiles, err := findAllJsonFiles(path)
		if err != nil {
			return fmt.Errorf("find json files in plugin %s: %v", name, err)
		}
		logger.Info("found json files in plugin",
			"plugin", name,
			"path", path,
			"total_files", len(jsonFiles),
		)
		for _, jsonFile := range jsonFiles {
			content, err := os.ReadFile(jsonFile)
			if err != nil {
				return fmt.Errorf("read file %s: %v", jsonFile, err)
			}

			var data dataGen.JsonLangData
			if err := json.Unmarshal(content, &data); err != nil {
				return fmt.Errorf("unmarshal json file %s: %v", jsonFile, err)
			}

			err = l.repo.Save(m.ToHash(data.Triggers), m.ToHash(data.Results))
			if err != nil {
				return fmt.Errorf("save plugin %s: %v", name, err)
			}
			logger.Info("parse plugin",
				"plugin", name,
				"triggers", data.Triggers,
				"results", data.Results,
			)
		}
	}

	return nil
}
func NewRegistry(repo Repo) Registry {
	return &langRegistry{
		repo: repo,
	}
}

// findAllJsonFiles 遍历目录，获取所有 .json 文件
func findAllJsonFiles(dir string) ([]string, error) {
	var jsonFiles []string

	// 读取目录所有文件
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// 筛选 .json 后缀文件
	for _, file := range files {
		// 跳过子目录
		if file.IsDir() {
			continue
		}

		// 只处理 .json 文件
		if strings.ToLower(filepath.Ext(file.Name())) == ".json" {
			fullPath := filepath.Join(dir, file.Name())
			jsonFiles = append(jsonFiles, fullPath)
		}
	}

	return jsonFiles, nil
}
