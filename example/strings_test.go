package example

import (
	"strings"
	"testing"
)

var (
	shortString  = "hello world"
	mediumString = strings.Repeat("hello world ", 10)
	longString   = strings.Repeat("hello world ", 100)
)

func BenchmarkReverseShort(b *testing.B) {
	for b.Loop() {
		Reverse(shortString)
	}
}

func BenchmarkReverseMedium(b *testing.B) {
	for b.Loop() {
		Reverse(mediumString)
	}
}

func BenchmarkReverseLong(b *testing.B) {
	for b.Loop() {
		Reverse(longString)
	}
}

func BenchmarkCountWordsShort(b *testing.B) {
	for b.Loop() {
		CountWords(shortString)
	}
}

func BenchmarkCountWordsMedium(b *testing.B) {
	for b.Loop() {
		CountWords(mediumString)
	}
}

func BenchmarkCountWordsLong(b *testing.B) {
	for b.Loop() {
		CountWords(longString)
	}
}

func BenchmarkContainsShort(b *testing.B) {
	for b.Loop() {
		Contains(shortString, "world")
	}
}

func BenchmarkContainsLong(b *testing.B) {
	for b.Loop() {
		Contains(longString, "world")
	}
}
