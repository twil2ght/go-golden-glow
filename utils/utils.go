package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func TrimAll(str []string) []string {
	for i, s := range str {
		str[i] = strings.TrimSpace(s)
	}
	return str
}

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
	root := findProjectRoot(wd)
	RootDir = filepath.Join(root, "data")
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
