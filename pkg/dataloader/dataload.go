package dataloader

import (
	"encoding/json"
	"goldenglow/m"
	"goldenglow/pkg/datagen"
	"goldenglow/pkg/log"
	"os"
)

var (
	logger = log.Default()
)

type Repo interface {
	Save(t, r m.Hash)
}
type Interface interface {
	Load()
}
type loader struct {
	repo Repo
}

func (l *loader) Load(rootDir string) {
	jsonFiles := findAllJsonFiles(rootDir)
	tagMap := m.Map[int]{}

	for _, jsonFile := range jsonFiles {
		content, err := os.ReadFile(jsonFile)
		if err != nil {
			OnLoadingError(jsonFile)
			continue
		}

		var data datagen.JsonFormatData
		if err := json.Unmarshal(content, &data); err != nil {
			OnLoadingError(jsonFile)
			continue
		}
		tagMap[data.Tag]++

		l.repo.Save(m.ToHash(data.Triggers), m.ToHash(data.Results))
	}

	OnFinishedInfoSum(tagMap)
}
