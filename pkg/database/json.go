package database

import (
	"encoding/json"
	"errors"
	"goldenglow/m"
	"goldenglow/pkg/log"
	"goldenglow/utils"
	"os"
	"path/filepath"
	"sync"
)

var (
	defaultJSONHDataPath = filepath.Join(utils.RootDir, "archive/Data/json/hash_data.json")
)

var logger = log.Default()

type jsonRepository struct {
	HDataPath string
	HData     map[string]m.Hash
	mu        *sync.Mutex
}

func (j *jsonRepository) Get(_ string) (string, error) { return "", nil }

func (j *jsonRepository) Set(_, _ string) error { return nil }

func (j *jsonRepository) Shutdown() error {
	return j.Save()
}

func (j *jsonRepository) HSet(tag string, value m.Hash) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.HData[tag] = value
	return nil
}

func (j *jsonRepository) HGet(tag string) (m.Hash, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	copied := m.Hash{}
	if value, ok := j.HData[tag]; ok {
		for k := range value {
			copied[k] = struct{}{}
		}
		return copied, nil
	}
	return nil, errors.New("not found")
}

func (j *jsonRepository) Init() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	file, err := os.OpenFile(j.HDataPath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = file.Close()
	}()

	stat, err := file.Stat()
	if err != nil {
		panic(err)
	}
	if stat.Size() == 0 {
		return nil
	}

	return json.NewDecoder(file).Decode(&j.HData)
}

func (j *jsonRepository) Save() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if err := SaveAsJson(j.HDataPath, j.HData); err != nil {
		return err
	}
	return nil
}
func SaveAsJson(path string, Data any) error {
	data, err := json.MarshalIndent(Data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
func NewJSONRepo(HDataPath string) Repository {
	repo := &jsonRepository{
		HDataPath: HDataPath,
		HData:     make(map[string]m.Hash),
		mu:        &sync.Mutex{},
	}
	return repo
}
