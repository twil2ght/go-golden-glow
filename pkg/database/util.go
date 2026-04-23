package database

import (
	"encoding/json"
	"os"
)

func SaveAsJson(path string, Data any) error {
	data, err := json.MarshalIndent(Data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
