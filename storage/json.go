package storage

import (
	"encoding/json"
	"goldenglow/m"
	"goldenglow/pkg/log"
	"os"
)

const (
	DefaultJSONPathRoot  = "./archive/Data/json"
	defaultJSONHDataPath = DefaultJSONPathRoot + "/hash_data.json"
	defaultJSONDataPath  = DefaultJSONPathRoot + "/data.json"
)

type jsonRepository struct {
	DataPath  string
	HDataPath string
	HData     map[string]m.Hash
	Data      map[string]string
}

func (j *jsonRepository) Shutdown() error {
	return j.Save()
}

func (j *jsonRepository) Get(key string) (string, error) {
	if value, ok := j.Data[key]; ok {
		return value, nil
	}
	return "", log.NotFound(key)
}

func (j *jsonRepository) Set(key, value string) error {
	j.Data[key] = value
	return nil
}

func (j *jsonRepository) HSet(tag string, value m.Hash) error {
	j.HData[tag] = value
	return nil
}

func (j *jsonRepository) HGet(tag string) (m.Hash, error) {
	return j.HData[tag], nil
}

func (j *jsonRepository) Init() error {
	j.HData = make(map[string]m.Hash)

	err := os.MkdirAll(DefaultJSONPathRoot, 0755)
	if err != nil {
		return err
	}

	_, err = os.Stat(j.HDataPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	file, err := os.ReadFile(j.HDataPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(file, &j.HData)
}

func (j *jsonRepository) Save() error {
	err := os.MkdirAll(DefaultJSONPathRoot, 0755)
	if err != nil {
		return err
	}
	if err := Save(j.HDataPath, j.HData); err != nil {
		return err
	}
	if err := Save(j.DataPath, j.Data); err != nil {
		return err
	}
	return nil
}
func Save(path string, Data any) error {
	data, err := json.MarshalIndent(Data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
func NewJSONRepo(HDataPath, DataPath string) Repository {
	if HDataPath == "" {
		HDataPath = defaultJSONHDataPath
	}
	if DataPath == "" {
		DataPath = defaultJSONDataPath
	}
	repo := &jsonRepository{
		HDataPath: HDataPath,
		DataPath:  DataPath,
		Data:      make(map[string]string),
		HData:     make(map[string]m.Hash),
	}
	return repo
}
