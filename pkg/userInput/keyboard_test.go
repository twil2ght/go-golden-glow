package userInput

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestKeyboard_Start(t *testing.T) {
	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func(r *os.File) {
		err := r.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(r)
	defer func(w *os.File) {
		err := w.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(w)

	// Save original stdin and restore after test
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Create keyboard
	k := NewKeyboard()

	// Create context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start keyboard in a goroutine
	go k.Start(ctx)

	// Simulate user input
	_, err = w.WriteString("test input\n")
	if err != nil {
		t.Fatal(err)
	}

	// Wait a bit for the input to be processed
	time.Sleep(100 * time.Millisecond)

	// Cancel the context to stop the keyboard
	cancel()

	// Wait a bit for the goroutine to stop
	time.Sleep(50 * time.Millisecond)

	// Check that the input was added to the queue
	item, shutdown := k.queue.Get()
	if shutdown {
		t.Error("unexpected shutdown")
	}
	if item != "test input" {
		t.Errorf("expected 'test input', got '%s'", item)
	}
}
