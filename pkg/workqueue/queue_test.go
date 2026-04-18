package workqueue

import (
	"testing"
	"time"
)

func TestDefaultQueue_AddGet(t *testing.T) {
	q := New[int]()

	if got := q.Len(); got != 0 {
		t.Fatalf("expected initial length 0, got %d", got)
	}

	q.Add(10)
	q.Add(20)

	if got := q.Len(); got != 2 {
		t.Fatalf("expected length 2 after adding items, got %d", got)
	}

	item, shutdown := q.Get()
	if shutdown {
		t.Fatal("expected shutdown=false on first Get")
	}
	if item != 10 {
		t.Fatalf("expected first item 10, got %d", item)
	}

	item, shutdown = q.Get()
	if shutdown {
		t.Fatal("expected shutdown=false on second Get")
	}
	if item != 20 {
		t.Fatalf("expected second item 20, got %d", item)
	}

	if got := q.Len(); got != 0 {
		t.Fatalf("expected empty queue after consuming all items, got %d", got)
	}
}

func TestDefaultQueue_ShutdownUnblocksGet(t *testing.T) {
	q := New[int]()
	result := make(chan struct {
		item     int
		shutdown bool
	}, 1)

	go func() {
		item, shutdown := q.Get()
		result <- struct {
			item     int
			shutdown bool
		}{item, shutdown}
	}()

	select {
	case <-result:
		t.Fatal("expected Get to block until Shutdown is called")
	case <-time.After(20 * time.Millisecond):
	}

	q.Shutdown()

	select {
	case res := <-result:
		if !res.shutdown {
			t.Fatal("expected shutdown=true after Shutdown")
		}
		if res.item != 0 {
			t.Fatalf("expected zero item on shutdown, got %d", res.item)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected Get to return after Shutdown")
	}
}

func TestDefaultQueue_AddAfterShutdownIsIgnored(t *testing.T) {
	q := New[string]()
	q.Shutdown()
	q.Add("hello")

	if got := q.Len(); got != 0 {
		t.Fatalf("expected length 0 after adding to shutdown queue, got %d", got)
	}

	item, shutdown := q.Get()
	if !shutdown {
		t.Fatal("expected shutdown=true after Get on shutdown queue")
	}
	if item != "" {
		t.Fatalf("expected empty string item on shutdown, got %q", item)
	}
}
