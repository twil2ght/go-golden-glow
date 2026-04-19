package datagen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewData(t *testing.T) {
	t.Run("creates data with correct fields", func(t *testing.T) {
		tData := []string{"trigger1", "trigger2"}
		rData := []string{"result1"}
		params := map[string]string{"key1": "value1"}
		extraType := AsExecutor

		data := NewData(tData, rData, params, extraType)

		if len(data.T()) != 2 || data.T()[0] != "trigger1" {
			t.Errorf("Expected T() to return %v, got %v", tData, data.T())
		}
		if len(data.R()) != 1 || data.R()[0] != "result1" {
			t.Errorf("Expected R() to return %v, got %v", rData, data.R())
		}
	})
}

func TestDataBuildExtra(t *testing.T) {
	tests := []struct {
		name      string
		t         []string
		r         []string
		params    map[string]string
		extraType DataExtraType
		namespace string
		expectedT []string
		expectedR []string
	}{
		{
			name:      "AsExecutor adds to R",
			t:         []string{"t1"},
			r:         []string{"r1"},
			params:    map[string]string{"p1": "v1"},
			extraType: AsExecutor,
			namespace: "test",
			expectedT: []string{"t1"},
			expectedR: []string{"r1", "[node:executor] [namespace:test] [p1:v1]"},
		},
		{
			name:      "AsChecker adds to T",
			t:         []string{"t1"},
			r:         []string{"r1"},
			params:    map[string]string{"p1": "v1", "p2": "v2"},
			extraType: AsChecker,
			namespace: "test",
			expectedT: []string{"t1", "[node:checker] [namespace:test] [p1:v1] [p2:v2]"},
			expectedR: []string{"r1"},
		},
		{
			name:      "AsExtractor adds to T",
			t:         []string{"t1"},
			r:         []string{"r1"},
			params:    map[string]string{},
			extraType: AsExtractor,
			namespace: "test",
			expectedT: []string{"t1", "[node:extractor] [namespace:test]"},
			expectedR: []string{"r1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := NewData(tt.t, tt.r, tt.params, tt.extraType)
			data.BuildExtra(tt.namespace)

			if len(data.T()) != len(tt.expectedT) {
				t.Errorf("Expected T length %d, got %d", len(tt.expectedT), len(data.T()))
			}
			for i, v := range data.T() {
				if i >= len(tt.expectedT) || v != tt.expectedT[i] {
					t.Errorf("Expected T[%d] = %s, got %s", i, tt.expectedT[i], v)
				}
			}

			if len(data.R()) != len(tt.expectedR) {
				t.Errorf("Expected R length %d, got %d", len(tt.expectedR), len(data.R()))
			}
			for i, v := range data.R() {
				if i >= len(tt.expectedR) || v != tt.expectedR[i] {
					t.Errorf("Expected R[%d] = %s, got %s", i, tt.expectedR[i], v)
				}
			}
		})
	}
}

func TestNewProvider(t *testing.T) {
	provider := NewProvider()
	if provider == nil {
		t.Error("NewProvider returned nil")
	}
}

func TestProviderAddAndRun(t *testing.T) {
	// Create a temp dir for testing
	tempDir, err := os.MkdirTemp("", "datagen_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Override RootDir for testing
	originalRootDir := RootDir
	RootDir = tempDir
	defer func() { RootDir = originalRootDir }()

	provider := NewProvider()
	data := NewData([]string{"t1"}, []string{"r1"}, map[string]string{"p1": "v1"}, AsExecutor)
	provider.Add("testdata", data)

	provider.Run("testns")

	// Check if file was created
	expectedPath := filepath.Join(tempDir, "testns", "testdata.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", expectedPath)
	}

	// Check content
	file, err := os.Open(expectedPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	var jsonData JsonFormatData
	if err := json.NewDecoder(file).Decode(&jsonData); err != nil {
		t.Fatal(err)
	}

	expected := JsonFormatData{
		Triggers: []string{"t1"},
		Results:  []string{"r1", "[node:executor] [namespace:testns] [p1:v1]"},
		Tag:      "testns",
	}

	if jsonData.Tag != expected.Tag || len(jsonData.Triggers) != len(expected.Triggers) || len(jsonData.Results) != len(expected.Results) {
		t.Errorf("Expected %v, got %v", expected, jsonData)
	}
}

func TestNewGenerator(t *testing.T) {
	generator := NewGenerator()
	if generator == nil {
		t.Error("NewGenerator returned nil")
	}
}

func TestGeneratorAddProviderAndRun(t *testing.T) {
	// Similar to provider test, but for generator
	tempDir, err := os.MkdirTemp("", "datagen_test_gen")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	originalRootDir := RootDir
	RootDir = tempDir
	defer func() { RootDir = originalRootDir }()

	generator := NewGenerator()
	provider := NewProvider()
	data := NewData([]string{"t1"}, []string{"r1"}, nil, AsChecker)
	provider.Add("data1", data)
	generator.AddProvider("ns1", provider)

	generator.Run()

	expectedPath := filepath.Join(tempDir, "ns1", "data1.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", expectedPath)
	}
}

func TestMakePath(t *testing.T) {
	tests := []struct {
		namespace string
		filename  string
		expected  string
	}{
		{"ns", "file", filepath.Join(RootDir, "ns", "file.json")},
		{"ns", "file.json", filepath.Join(RootDir, "ns", "file.json")},
		{"", "file", ""},
		{"ns", "", ""},
	}

	for _, tt := range tests {
		result := makePath(tt.namespace, tt.filename)
		if result != tt.expected {
			t.Errorf("makePath(%s, %s) = %s, expected %s", tt.namespace, tt.filename, result, tt.expected)
		}
	}
}
