package userInput

import (
	"encoding/json"
	"goldenglow/pkg/log"
	"goldenglow/pkg/messageQueue"
	"goldenglow/utils"
	"os"
	"path/filepath"
	"strings"
)

var (
	logger = log.Default()
)

type File struct {
	queue  messageQueue.Interface
	loaded map[string]bool
	Count  int
}

type ValidFormat struct {
	Commands []string `json:"commands"`
}

type Data struct {
	Data []ValidFormat `json:"data"`
}

type fileFormat struct {
	Dependencies []string      `json:"dependencies"`
	Data         []ValidFormat `json:"data"`
}

func (f *File) Run(dir string) {
	if f.loaded == nil {
		f.loaded = make(map[string]bool)
	}
	jsonFiles := utils.FindAllJsonFiles(dir)
	logger.Debug("userInput file starting", "dir", dir, "amount", len(jsonFiles))
	for _, jsonFile := range jsonFiles {
		f.loadFile(jsonFile)
	}
}

func (f *File) loadFile(jsonFile string) {
	if f.loaded[jsonFile] {
		return
	}
	f.loaded[jsonFile] = true

	content, err := os.ReadFile(jsonFile)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	var ff fileFormat
	if err := json.Unmarshal(content, &ff); err != nil {
		logger.Error(err.Error())
		return
	}

	for _, item := range ff.Data {
		for _, command := range item.Commands {
			f.queue.Add(command)
			f.Count++
		}
	}

	dir := filepath.Dir(jsonFile)
	for _, dep := range ff.Dependencies {
		if strings.HasSuffix(dep, "_test") {
			continue
		}
		depPath := filepath.Join(dir, dep+".json")
		f.loadFile(depPath)
	}
}

func NewFile(queue messageQueue.Interface) *File {
	return &File{
		queue:  queue,
		loaded: make(map[string]bool),
	}
}
