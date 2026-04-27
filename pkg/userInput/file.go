package userInput

import (
	"encoding/json"
	"goldenglow/pkg/log"
	"goldenglow/pkg/messageQueue"
	"goldenglow/utils"
	"os"
)

var (
	logger = log.Default()
)

type File struct {
	queue messageQueue.Interface
}
type ValidFormat struct {
	Commands []string `json:"commands"`
}

func (f *File) Run(dir string) {
	jsonFiles := utils.FindAllJsonFiles(dir)
	logger.Debug("userInput file starting ", "dir", dir, "amount", len(jsonFiles))
	for _, jsonFile := range jsonFiles {
		content, err := os.ReadFile(jsonFile)
		if err != nil {
			logger.Error(err.Error())
			continue
		}
		// First try unmarshaling as an array of ValidFormat
		var dataList []ValidFormat
		err = json.Unmarshal(content, &dataList)
		if err == nil && len(dataList) > 0 {
			for _, item := range dataList {
				for _, command := range item.Commands {
					f.queue.Add(command)
				}
			}
			continue
		}
		// Fall back to a single ValidFormat
		data := ValidFormat{}
		err = json.Unmarshal(content, &data)
		if err != nil {
			logger.Error(err.Error())
			continue
		}
		for _, command := range data.Commands {
			f.queue.Add(command)
		}
	}
}

func NewFile(queue messageQueue.Interface) *File {
	return &File{
		queue: queue,
	}
}
