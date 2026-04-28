package database

import (
	"encoding/json"
	"errors"
	"goldenglow/m"
	"goldenglow/utils"
	"os"
	"path/filepath"
	"sync"
)

var (
	defaultJSONHDataPath = filepath.Join(utils.RootDir, "archive/Data/json/hash_data.json")
)

type JSON[T any] struct {
	path string
	data map[string]T
	mu   *sync.RWMutex
}

func (J *JSON[T]) Save() error {
	J.mu.Lock()
	defer J.mu.Unlock()
	return SaveAsJson(J.path, J.data)
}

func (J *JSON[T]) Load() error {
	J.mu.Lock()
	defer J.mu.Unlock()
	file, err := os.OpenFile(J.path, os.O_RDONLY|os.O_CREATE, 0644)
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

	return json.NewDecoder(file).Decode(&J.data)
}
func NewJSON[T any](path string) JSON[T] {
	return JSON[T]{
		path: path,
		data: make(map[string]T),
		mu:   &sync.RWMutex{},
	}
}

type jsonRepository struct {
	JSON[m.Hash]
}

func (j *jsonRepository) Get(_ string) (string, error) { return "", nil }

func (j *jsonRepository) Set(_, _ string) error { return nil }

func (j *jsonRepository) HSet(tag string, value m.Hash) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.data[tag] = value
	return nil
}

func (j *jsonRepository) HGet(tag string) (m.Hash, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	copied := m.Hash{}
	if value, ok := j.data[tag]; ok {
		for k := range value {
			copied[k] = struct{}{}
		}
		return copied, nil
	}
	return nil, errors.New("not found")
}

func (j *jsonRepository) HDel(tag string, subKeys ...string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if _, ok := j.data[tag]; !ok {
		return
	}
	if len(subKeys) == 0 {
		delete(j.data, tag)
		return
	}
	for _, subKey := range subKeys {
		delete(j.data[tag], subKey)
	}
	if len(j.data[tag]) == 0 {
		delete(j.data, tag)
	}
}

func (j *jsonRepository) Init() error {
	return j.Load()
}
func (j *jsonRepository) Shutdown() error {
	return j.Save()
}
func NewJSONRepo(path string) Repository {
	repo := &jsonRepository{
		JSON: NewJSON[m.Hash](path),
	}
	return repo
}
