package userInput

import (
	"encoding/json"
	"goldenglow/pkg/messageQueue"
	"os"
	"path/filepath"
	"testing"
)

func TestNewFile(t *testing.T) {
	q := messageQueue.New("test.json")
	f := NewFile(q)
	if f == nil {
		t.Fatal("expected NewFile to return a non-nil File")
	}
}

func TestFile_Run_AddsCommandsFromValidJsonFiles(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "input.json")
	data := Data{Data: []ValidFormat{{Commands: []string{"foo", "bar"}}}}
	content, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	q := messageQueue.New("test.json")
	f := NewFile(q)
	f.Run(dir)

	item, shutdown := q.Get()
	if shutdown {
		t.Fatal("unexpected shutdown")
	}
	if item != "foo" {
		t.Fatalf("expected first command 'foo', got '%s'", item)
	}

	item, shutdown = q.Get()
	if shutdown {
		t.Fatal("unexpected shutdown")
	}
	if item != "bar" {
		t.Fatalf("expected second command 'bar', got '%s'", item)
	}

	if q.Len() != 0 {
		t.Fatalf("expected queue to be empty after consuming all items, got len=%d", q.Len())
	}
}

func TestFile_Run_LogsErrorForInvalidJsonButContinues(t *testing.T) {
	dir := t.TempDir()
	validPath := filepath.Join(dir, "valid.json")
	invalidPath := filepath.Join(dir, "invalid.json")
	validData := Data{Data: []ValidFormat{{Commands: []string{"hello"}}}}
	validContent, err := json.Marshal(validData)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(validPath, validContent, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(invalidPath, []byte("{not json}"), 0o644); err != nil {
		t.Fatal(err)
	}

	q := messageQueue.New("test.json")
	f := NewFile(q)
	f.Run(dir)

	item, shutdown := q.Get()
	if shutdown {
		t.Fatal("unexpected shutdown")
	}
	if item != "hello" {
		t.Fatalf("expected command 'hello' after invalid file, got '%s'", item)
	}
	if q.Len() != 0 {
		t.Fatalf("expected queue to be empty after consuming valid item, got len=%d", q.Len())
	}
}
