package example

import (
	"math/rand"
	"testing"
)

var (
	smallSlice  = generateRandomSlice(100)
	mediumSlice = generateRandomSlice(1000)
	largeSlice  = generateRandomSlice(10000)
)

func generateRandomSlice(n int) []int {
	slice := make([]int, n)
	for i := range n {
		slice[i] = rand.Intn(10000)
	}
	return slice
}

// BubbleSort benchmarks
func BenchmarkBubbleSortSmall(b *testing.B) {
	for b.Loop() {
		BubbleSort(smallSlice)
	}
}

func BenchmarkBubbleSortMedium(b *testing.B) {
	for b.Loop() {
		BubbleSort(mediumSlice)
	}
}

func BenchmarkBubbleSortLarge(b *testing.B) {
	for b.Loop() {
		BubbleSort(largeSlice)
	}
}

// SelectionSort benchmarks
func BenchmarkSelectionSortSmall(b *testing.B) {
	for b.Loop() {
		SelectionSort(smallSlice)
	}
}

func BenchmarkSelectionSortMedium(b *testing.B) {
	for b.Loop() {
		SelectionSort(mediumSlice)
	}
}

func BenchmarkSelectionSortLarge(b *testing.B) {
	for b.Loop() {
		SelectionSort(largeSlice)
	}
}

// InsertionSort benchmarks
func BenchmarkInsertionSortSmall(b *testing.B) {
	for b.Loop() {
		InsertionSort(smallSlice)
	}
}

func BenchmarkInsertionSortMedium(b *testing.B) {
	for b.Loop() {
		InsertionSort(mediumSlice)
	}
}

func BenchmarkInsertionSortLarge(b *testing.B) {
	for b.Loop() {
		InsertionSort(largeSlice)
	}
}

// MergeSort benchmarks
func BenchmarkMergeSortSmall(b *testing.B) {
	for b.Loop() {
		MergeSort(smallSlice)
	}
}

func BenchmarkMergeSortMedium(b *testing.B) {
	for b.Loop() {
		MergeSort(mediumSlice)
	}
}

func BenchmarkMergeSortLarge(b *testing.B) {
	for b.Loop() {
		MergeSort(largeSlice)
	}
}

// QuickSort benchmarks
func BenchmarkQuickSortSmall(b *testing.B) {
	for b.Loop() {
		QuickSort(smallSlice)
	}
}

func BenchmarkQuickSortMedium(b *testing.B) {
	for b.Loop() {
		QuickSort(mediumSlice)
	}
}

func BenchmarkQuickSortLarge(b *testing.B) {
	for b.Loop() {
		QuickSort(largeSlice)
	}
}

// HeapSort benchmarks
func BenchmarkHeapSortSmall(b *testing.B) {
	for b.Loop() {
		HeapSort(smallSlice)
	}
}

func BenchmarkHeapSortMedium(b *testing.B) {
	for b.Loop() {
		HeapSort(mediumSlice)
	}
}

func BenchmarkHeapSortLarge(b *testing.B) {
	for b.Loop() {
		HeapSort(largeSlice)
	}
}

// ShellSort benchmarks
func BenchmarkShellSortSmall(b *testing.B) {
	for b.Loop() {
		ShellSort(smallSlice)
	}
}

func BenchmarkShellSortMedium(b *testing.B) {
	for b.Loop() {
		ShellSort(mediumSlice)
	}
}

func BenchmarkShellSortLarge(b *testing.B) {
	for b.Loop() {
		ShellSort(largeSlice)
	}
}

// CombSort benchmarks
func BenchmarkCombSortSmall(b *testing.B) {
	for b.Loop() {
		CombSort(smallSlice)
	}
}

func BenchmarkCombSortMedium(b *testing.B) {
	for b.Loop() {
		CombSort(mediumSlice)
	}
}

func BenchmarkCombSortLarge(b *testing.B) {
	for b.Loop() {
		CombSort(largeSlice)
	}
}

// CocktailSort benchmarks
func BenchmarkCocktailSortSmall(b *testing.B) {
	for b.Loop() {
		CocktailSort(smallSlice)
	}
}

func BenchmarkCocktailSortMedium(b *testing.B) {
	for b.Loop() {
		CocktailSort(mediumSlice)
	}
}

func BenchmarkCocktailSortLarge(b *testing.B) {
	for b.Loop() {
		CocktailSort(largeSlice)
	}
}

// GnomeSort benchmarks
func BenchmarkGnomeSortSmall(b *testing.B) {
	for b.Loop() {
		GnomeSort(smallSlice)
	}
}

func BenchmarkGnomeSortMedium(b *testing.B) {
	for b.Loop() {
		GnomeSort(mediumSlice)
	}
}

func BenchmarkGnomeSortLarge(b *testing.B) {
	for b.Loop() {
		GnomeSort(largeSlice)
	}
}

// StdSort benchmarks
func BenchmarkStdSortSmall(b *testing.B) {
	for b.Loop() {
		StdSort(smallSlice)
	}
}

func BenchmarkStdSortMedium(b *testing.B) {
	for b.Loop() {
		StdSort(mediumSlice)
	}
}

func BenchmarkStdSortLarge(b *testing.B) {
	for b.Loop() {
		StdSort(largeSlice)
	}
}

// RadixSort benchmarks
func BenchmarkRadixSortSmall(b *testing.B) {
	for b.Loop() {
		RadixSort(smallSlice)
	}
}

func BenchmarkRadixSortMedium(b *testing.B) {
	for b.Loop() {
		RadixSort(mediumSlice)
	}
}

func BenchmarkRadixSortLarge(b *testing.B) {
	for b.Loop() {
		RadixSort(largeSlice)
	}
}
