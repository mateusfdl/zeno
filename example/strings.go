package example

import (
	"strings"
)

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func CountWords(s string) int {
	return len(strings.Fields(s))
}

func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
