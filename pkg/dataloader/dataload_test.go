package dataloader

import (
	"encoding/json"
	"fmt"
	"goldenglow/m"
	"goldenglow/pkg/database"
	"goldenglow/pkg/datagen"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// mockRepo implements the Repo interface for testing
type mockRepo struct {
	savedData []savedItem
}

func (mr *mockRepo) GetRepo() database.database {
	//TODO implement me
	panic("implement me")
}

type savedItem struct {
	triggers m.Hash
	results  m.Hash
}

func (mr *mockRepo) Save(t, r m.Hash) {
	mr.savedData = append(mr.savedData, savedItem{triggers: t, results: r})
}

func mapsEqual(a, b m.Hash) bool {
	return reflect.DeepEqual(a, b)
}

func TestLoader_Load(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "dataloader_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			fmt.Printf("Failed to remove temp dir: %v", err)
		}
	}(tempDir)

	// Create sample JSON files
	sampleData1 := datagen.JsonFormatData{
		Triggers: []string{"hello", "hi"},
		Results:  []string{"greeting"},
		Tag:      "greeting",
	}
	sampleData2 := datagen.JsonFormatData{
		Triggers: []string{"bye", "goodbye"},
		Results:  []string{"farewell"},
		Tag:      "farewell",
	}
	sampleData3 := datagen.JsonFormatData{
		Triggers: []string{"thanks"},
		Results:  []string{"welcome"},
		Tag:      "greeting", // Same tag as sampleData1
	}

	files := []datagen.JsonFormatData{sampleData1, sampleData2, sampleData3}
	for i, data := range files {
		filePath := filepath.Join(tempDir, fmt.Sprintf("test%d.json", i))
		content, err := json.Marshal(data)
		if err != nil {
			t.Fatalf("Failed to marshal sample data: %v", err)
		}
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			t.Fatalf("Failed to write sample file: %v", err)
		}
	}

	// Create mock repo
	mockRepo := &mockRepo{}

	// Create loader
	loader := New(mockRepo)

	// Load data
	loader.Load(tempDir)

	// Verify saved data
	expectedSaves := 3
	if len(mockRepo.savedData) != expectedSaves {
		t.Errorf("Expected %d saves, got %d", expectedSaves, len(mockRepo.savedData))
	}

	// Check specific saves
	expected := []savedItem{
		{triggers: m.ToHash([]string{"hello", "hi"}), results: m.ToHash([]string{"greeting"})},
		{triggers: m.ToHash([]string{"bye", "goodbye"}), results: m.ToHash([]string{"farewell"})},
		{triggers: m.ToHash([]string{"thanks"}), results: m.ToHash([]string{"welcome"})},
	}

	for i, exp := range expected {
		if i >= len(mockRepo.savedData) {
			t.Errorf("Missing save at index %d", i)
			continue
		}
		actual := mockRepo.savedData[i]
		if !mapsEqual(actual.triggers, exp.triggers) {
			t.Errorf("Save %d: expected triggers %v, got %v", i, exp.triggers, actual.triggers)
		}
		if !mapsEqual(actual.results, exp.results) {
			t.Errorf("Save %d: expected results %v, got %v", i, exp.results, actual.results)
		}
	}
}

func TestLoader_Load_InvalidJson(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "dataloader_test_invalid")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			fmt.Printf("Failed to remove temp dir: %v", err)
		}
	}(tempDir)

	// Create an invalid JSON file
	invalidFilePath := filepath.Join(tempDir, "invalid.json")
	if err := os.WriteFile(invalidFilePath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid file: %v", err)
	}

	// Create mock repo
	mockRepo := &mockRepo{}

	// Create loader
	loader := New(mockRepo)

	// Load data - should not crash, just log error
	loader.Load(tempDir)

	// Should have no saves since JSON is invalid
	if len(mockRepo.savedData) != 0 {
		t.Errorf("Expected 0 saves for invalid JSON, got %d", len(mockRepo.savedData))
	}
}

func TestLoader_Load_NoJsonFiles(t *testing.T) {
	// Create a temporary directory with no JSON files
	tempDir, err := os.MkdirTemp("", "dataloader_test_empty")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			fmt.Printf("Failed to remove temp dir: %v", err)
		}
	}(tempDir)

	// Create a non-JSON file
	nonJsonFilePath := filepath.Join(tempDir, "text.txt")
	if err := os.WriteFile(nonJsonFilePath, []byte("some text"), 0644); err != nil {
		t.Fatalf("Failed to write non-JSON file: %v", err)
	}

	// Create mock repo
	mockRepo := &mockRepo{}

	// Create loader
	loader := New(mockRepo)

	// Load data
	loader.Load(tempDir)

	// Should have no saves
	if len(mockRepo.savedData) != 0 {
		t.Errorf("Expected 0 saves for no JSON files, got %d", len(mockRepo.savedData))
	}
}
