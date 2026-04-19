package storage

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"goldenglow/m"
)

func TestJSONRepo_InitCreatesFilesAndLoads(t *testing.T) {
	root := t.TempDir()
	prevRoot := DefaultJSONPathRoot
	DefaultJSONPathRoot = root
	defer func() { DefaultJSONPathRoot = prevRoot }()

	hDataPath := filepath.Join(root, "hash_data.json")
	dataPath := filepath.Join(root, "data.json")
	repo := NewJSONRepo(hDataPath, dataPath).(*jsonRepository)

	if err := repo.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if _, err := os.Stat(hDataPath); err != nil {
		t.Fatalf("hash data file not created: %v", err)
	}
	if _, err := os.Stat(dataPath); err != nil {
		t.Fatalf("data file not created: %v", err)
	}

	if len(repo.Data) != 0 {
		t.Fatalf("expected empty Data map after init, got %d entries", len(repo.Data))
	}
	if len(repo.HData) != 0 {
		t.Fatalf("expected empty HData map after init, got %d entries", len(repo.HData))
	}
}

func TestJSONRepo_GetSetAndPersistence(t *testing.T) {
	root := t.TempDir()
	prevRoot := DefaultJSONPathRoot
	DefaultJSONPathRoot = root
	defer func() { DefaultJSONPathRoot = prevRoot }()

	hDataPath := filepath.Join(root, "hash_data.json")
	dataPath := filepath.Join(root, "data.json")
	repo := NewJSONRepo(hDataPath, dataPath).(*jsonRepository)

	if err := repo.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if err := repo.Set("key1", "value1"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	hashValue := m.ToHash([]string{"item1", "item2"})
	if err := repo.HSet("tag1", hashValue); err != nil {
		t.Fatalf("HSet() error = %v", err)
	}

	if err := repo.Shutdown(); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	newRepo := NewJSONRepo(hDataPath, dataPath)
	if err := newRepo.Init(); err != nil {
		t.Fatalf("second Init() error = %v", err)
	}

	value, err := newRepo.Get("key1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if value != "value1" {
		t.Fatalf("expected value1, got %q", value)
	}

	loadedHash, err := newRepo.HGet("tag1")
	if err != nil {
		t.Fatalf("HGet() error = %v", err)
	}
	if !reflect.DeepEqual(loadedHash, hashValue) {
		t.Fatalf("expected HGet to return %v, got %v", hashValue, loadedHash)
	}
}

func TestJSONRepo_GetMissingKeyReturnsError(t *testing.T) {
	root := t.TempDir()
	prevRoot := DefaultJSONPathRoot
	DefaultJSONPathRoot = root
	defer func() { DefaultJSONPathRoot = prevRoot }()

	hDataPath := filepath.Join(root, "hash_data.json")
	dataPath := filepath.Join(root, "data.json")
	repo := NewJSONRepo(hDataPath, dataPath).(*jsonRepository)

	if err := repo.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if _, err := repo.Get("missing-key"); err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
}

func TestJSONRepo_HGetMissingTagReturnsNilHash(t *testing.T) {
	root := t.TempDir()
	prevRoot := DefaultJSONPathRoot
	DefaultJSONPathRoot = root
	defer func() { DefaultJSONPathRoot = prevRoot }()

	hDataPath := filepath.Join(root, "hash_data.json")
	dataPath := filepath.Join(root, "data.json")
	repo := NewJSONRepo(hDataPath, dataPath).(*jsonRepository)

	if err := repo.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	value, err := repo.HGet("missing-tag")
	if err != nil {
		t.Fatalf("HGet() error = %v", err)
	}
	if value != nil {
		t.Fatalf("expected nil hash for missing tag, got %v", value)
	}
}
