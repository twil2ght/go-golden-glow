package utils

import (
	"strings"
)

func TrimAll(str []string) []string {
	for i, s := range str {
		str[i] = strings.TrimSpace(s)
	}
	return str
}
