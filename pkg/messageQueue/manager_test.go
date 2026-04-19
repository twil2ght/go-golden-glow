package messageQueue

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestManager_Add(t *testing.T) {
	m := NewManager().(*manager)

	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)

	m.Add("provider1", ch1)
	m.Add("provider2", ch2)

	// Check that items are registered
	count := 0
	m.items.Range(func(key string, item chan string) bool {
		count++
		return true
	})

	if count != 2 {
		t.Errorf("Expected 2 providers, got %d", count)
	}
}

func TestManager_Start(t *testing.T) {
	m := NewManager()
	mq := New("test_cache").(*msgQueue)

	ch1 := make(chan string, 10)
	ch2 := make(chan string, 10)

	m.Add("provider1", ch1)
	m.Add("provider2", ch2)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.Start(mq, ctx)
	}()

	// Send some messages
	ch1 <- "msg1"
	ch2 <- "msg2"
	ch1 <- "msg3"

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	// Cancel context to stop
	cancel()
	wg.Wait()

	// Check that messages were added to queue
	if mq.Len() != 3 {
		t.Errorf("Expected 3 messages in queue, got %d", mq.Len())
	}

	// Get messages
	msg1, _ := mq.Get()
	msg2, _ := mq.Get()
	msg3, _ := mq.Get()

	messages := []string{msg1, msg2, msg3}
	expected := []string{"msg1", "msg2", "msg3"}

	for _, exp := range expected {
		found := false
		for _, msg := range messages {
			if msg == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected message '%s' not found in %v", exp, messages)
		}
	}
}

func TestManager_Start_WithCancel(t *testing.T) {
	m := NewManager()
	mq := New("test_cache").(*msgQueue)

	ch := make(chan string, 10)
	m.Add("provider", ch)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.Start(mq, ctx)
	}()

	// Send a message
	ch <- "test_msg"

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Cancel immediately
	cancel()
	wg.Wait()

	// Check if message was processed
	if mq.Len() == 0 {
		t.Error("Expected at least one message in queue")
	}
}
