package utils

import (
	"fmt"
	"goldenglow/m"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v3"
)

var (
	RootDir string
)

// init get project root and set the RootDir variable
func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	RootDir = findProjectRoot(wd)
}

// findProjectRoot find project root (where go.mod is)
func findProjectRoot(start string) string {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return start
}

// NotNull parameter must be passed in pairs: key1, val1, key2, val2...
// check val whether it is nil / empty string
func NotNull(args ...any) error {
	if len(args)%2 != 0 {
		return fmt.Errorf("args should be even,current amount: %d", len(args))
	}

	for i := 0; i < len(args); i += 2 {
		key := args[i]
		val := args[i+1]

		if isNil(val) {
			return fmt.Errorf("%s is nil", key)
		}

		if str, ok := val.(string); ok && str == "" {
			return fmt.Errorf("%s is empty", key)
		}
	}

	return nil
}

// isNil check if nil safely (compatible with error etc. interface types)
func isNil(a any) bool {
	if a == nil {
		return true
	}
	v := reflect.ValueOf(a)
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	default:
		return false
	}
}
func ReadConfig() m.Map[any] {
	var cfg m.Map[any]
	content, _ := os.ReadFile(filepath.Join(RootDir, "config.yaml"))
	err := yaml.Unmarshal(content, &cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}
