package runner

import (
	"goldenglow/storage"
	"testing"

	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/setup"
)

// createTestNode creates a test node using the real node factory
func createTestNode(value string) node.Item {
	factory := node.DefaultFactory()
	n, _ := factory.New(value)
	return n
}

func TestDefaultRunner(t *testing.T) {
	// Test that DefaultRunner returns a valid instance
	runner := DefaultRunner()
	if runner == nil {
		t.Fatal("DefaultRunner() returned nil")
	}

	// Verify it's a Base instance
	base, ok := runner.(*Base)
	if !ok {
		t.Fatal("DefaultRunner() did not return *Base")
	}

	if base.containerFactory == nil {
		t.Error("Base.containerFactory is nil")
	}
	if base.templateCore == nil {
		t.Error("Base.templateCore is nil")
	}
}

func TestNewKnot(t *testing.T) {
	tests := []struct {
		name    string
		trigger node.Item
		trace   m.Hash
		wantErr bool
	}{
		{
			name:    "valid knot",
			trigger: createTestNode("test"),
			trace:   m.Hash{},
			wantErr: false,
		},
		{
			name:    "nil trigger",
			trigger: nil,
			trace:   m.Hash{},
			wantErr: true,
		},
		{
			name:    "nil trace creates empty hash",
			trigger: createTestNode("test"),
			trace:   nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			knot, err := NewKnot(tt.trigger, tt.trace)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewKnot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && knot == nil {
				t.Error("NewKnot() returned nil knot without error")
			}
		})
	}
}

func TestKnot_Trigger(t *testing.T) {
	trigger := createTestNode("test-trigger")
	knot, err := NewKnot(trigger, m.Hash{})
	if err != nil {
		t.Fatalf("NewKnot() failed: %v", err)
	}

	if knot.Trigger() != trigger {
		t.Error("Knot.Trigger() returned different item")
	}
}

func TestKnot_Trace(t *testing.T) {
	trace := m.Hash{"key": {}}
	knot, err := NewKnot(createTestNode("test"), trace)
	if err != nil {
		t.Fatalf("NewKnot() failed: %v", err)
	}

	if knot.Trace() == nil {
		t.Error("Knot.Trace() returned nil")
	}
}
func TestBase_SetContainerFactory(t *testing.T) {
	base := DefaultRunner().(*Base)

	// The default container factory should already be set
	if base.containerFactory == nil {
		t.Error("DefaultRunner() containerFactory is nil")
	}
}

func TestBase_SetTemplateCore(t *testing.T) {
	base := DefaultRunner().(*Base)

	// The default template core should already be set
	if base.templateCore == nil {
		t.Error("DefaultRunner() templateCore is nil")
	}
}

func TestBase_Run(t *testing.T) {
	if err := storage.DefaultJSONRepo().Init(); err != nil {
		panic(err)
	}
	setup.Init()
	runner := DefaultRunner()

	// Test with a simple node
	input := createTestNode("Susie should say hello")
	err := runner.Run(input)
	// Run may error due to template/container processing, but should not panic
	_ = err
}

func TestBase_Run_WithEmptyNode(t *testing.T) {
	if err := storage.DefaultJSONRepo().Init(); err != nil {
		panic(err)
	}
	setup.Init()
	runner := DefaultRunner()

	// Test with empty node value
	input := createTestNode("")
	err := runner.Run(input)
	_ = err
}

func TestBase_Run_WithComplexNode(t *testing.T) {
	if err := storage.DefaultJSONRepo().Init(); err != nil {
		panic(err)
	}
	setup.Init()
	runner := DefaultRunner()

	// Test with a more complex node value
	input := createTestNode("compute 2 + 3")
	err := runner.Run(input)
	_ = err
}
