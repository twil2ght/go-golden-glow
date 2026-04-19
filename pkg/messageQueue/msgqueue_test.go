package messageQueue

import (
	"testing"
)

func TestMsgQueue_Add(t *testing.T) {
	mq := New("test_cache").(*msgQueue)

	// Test adding items
	mq.Add("message1")
	mq.Add("message2")

	// Check logs
	expectedLogs := []string{"message1", "message2"}
	if len(mq.logs) != len(expectedLogs) {
		t.Errorf("Expected %d logs, got %d", len(expectedLogs), len(mq.logs))
	}
	for i, expected := range expectedLogs {
		if i >= len(mq.logs) || mq.logs[i] != expected {
			t.Errorf("Log %d: expected %s, got %s", i, expected, mq.logs[i])
		}
	}

	// Check underlying queue
	if mq.Len() != 2 {
		t.Errorf("Expected queue length 2, got %d", mq.Len())
	}
}

func TestMsgQueue_Get(t *testing.T) {
	mq := New("test_cache").(*msgQueue)

	mq.Add("message1")
	mq.Add("message2")

	// Get items
	item1, shutdown1 := mq.Get()
	if shutdown1 {
		t.Error("Expected not shutdown")
	}
	if item1 != "message1" {
		t.Errorf("Expected 'message1', got '%s'", item1)
	}

	item2, shutdown2 := mq.Get()
	if shutdown2 {
		t.Error("Expected not shutdown")
	}
	if item2 != "message2" {
		t.Errorf("Expected 'message2', got '%s'", item2)
	}

	// Queue should be empty
	if mq.Len() != 0 {
		t.Errorf("Expected queue length 0, got %d", mq.Len())
	}
}

func TestMsgQueue_Shutdown(t *testing.T) {
	mq := New("test_cache").(*msgQueue)

	mq.Add("message1")
	mq.Shutdown()

	// After shutdown, Get should return shutdown=true
	_, shutdown := mq.Get()
	if !shutdown {
		t.Error("Expected shutdown=true after Shutdown()")
	}

	// Adding after shutdown should not add to queue
	mq.Add("message2")
	if mq.Len() != 0 {
		t.Errorf("Expected queue length 0 after adding post-shutdown, got %d", mq.Len())
	}
}

func TestMsgQueue_Save(t *testing.T) {
	mq := New("test_cache").(*msgQueue)

	// Save is currently empty, so just ensure it doesn't panic
	mq.Save()
}
