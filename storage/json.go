package storage

import (
	"encoding/json"
	"goldenglow/m"
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

func (j *JSONRepository) Get(key string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (j *JSONRepository) Set(key, value string) error {
	//TODO implement me
	panic("implement me")
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
	_ = os.MkdirAll(j.HDataPath, 0755)

	data, err := json.MarshalIndent(j.HData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(j.HDataPath, data, 0644)
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
