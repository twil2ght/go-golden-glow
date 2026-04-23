package dataloader

import (
	"encoding/json"
	"goldenglow/m"
	"goldenglow/pkg/brainsaver"
	"goldenglow/pkg/datagen"
	"goldenglow/pkg/log"
	"goldenglow/utils"
	"os"
)

var (
	logger = log.Default()
)

type Repo interface {
	Save(t, r m.Hash)
}
type Interface interface {
	Load(rootDir string)
}
type loader struct {
	repo Repo
}

func (l *loader) Load(rootDir string) {
	jsonFiles := utils.FindAllJsonFiles(rootDir)
	tagMap := m.Map[int]{}
	OnStart(rootDir, len(jsonFiles))

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
		logger.Info("Loading data:", "tag", data.Tag, "T", data.Triggers, "R", data.Results)
		l.repo.Save(m.ToHash(data.Triggers), m.ToHash(data.Results))
	}

	OnFinishedInfoSum(tagMap)
}
func New(repo Repo) Interface {
	return &loader{
		repo: repo,
	}
}
func Default() Interface {
	return New(brainsaver.DefaultService())
}
