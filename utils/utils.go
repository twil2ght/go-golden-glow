package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

var (
	RootDir string
)

func init() {
	// 获取项目根目录（自动识别，永远正确）
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// 向上查找找到 go.mod 所在目录 = 项目根
	RootDir = findProjectRoot(wd)
}

// 自动找项目根（有 go.mod 就是根）
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

// NotNull 参数必须成对传入：key1, val1, key2, val2...
// 只校验 val 是否为 nil / 空字符串
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

// isNil 安全判断 nil（兼容 error 等接口类型）
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
