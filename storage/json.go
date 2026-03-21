package storage

import (
	"encoding/json"
	"goldenglow/m"
	"goldenglow/pkg/log"
	"os"
)

const (
	DefaultJSONPathRoot  = "./archive/HData/json"
	DefaultJSONHDataPath = DefaultJSONPathRoot + "/HData.json"
	DefaultJSONDataPath  = DefaultJSONPathRoot + "/Data.json"
)

type JSONRepository struct {
	DataPath  string
	HDataPath string
	HData     map[string]m.Hash
	Data      map[string]string
}

func (j *JSONRepository) Shutdown() error {
	return j.Save()
}

func (j *JSONRepository) Get(key string) (string, error) {
	if value, ok := j.Data[key]; ok {
		return value, nil
	}
	return "", log.NotFound(key)
}

func (j *JSONRepository) Set(key, value string) error {
	j.Data[key] = value
	return nil
}

func (j *JSONRepository) HSet(tag string, value m.Hash) error {
	j.HData[tag] = value
	return nil
}

func (j *JSONRepository) HGet(tag string) (m.Hash, error) {
	return j.HData[tag], nil
}

func (j *JSONRepository) Init() {
	j.HData = make(map[string]m.Hash)
	// 如果文件不存在，直接返回（后续Save会创建）
	if _, err := os.Stat(j.HDataPath); os.IsNotExist(err) {
		return
	}
	file, _ := os.ReadFile(j.HDataPath)
	_ = json.Unmarshal(file, &j.HData)
}

func (j *JSONRepository) Save() error {
	if err := Save(j.HDataPath, j.HData); err != nil {
		return err
	}
	if err := Save(j.DataPath, j.Data); err != nil {
		return err
	}
	return nil
}
func Save(path string, Data any) error {
	_ = os.MkdirAll(path, 0755)
	data, err := json.MarshalIndent(Data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
func NewJSONRepo(HDataPath, DataPath string) HashRepository {
	if HDataPath == "" {
		HDataPath = DefaultJSONHDataPath
	}
	if DataPath == "" {
		DataPath = DefaultJSONDataPath
	}
	repo := &JSONRepository{
		HDataPath: HDataPath,
		DataPath:  DataPath,
	}
	repo.Init()
	return repo
}
