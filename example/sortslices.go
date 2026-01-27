package example

import (
	"sort"
)

func BubbleSort(arr []int) []int {
	n := len(arr)
	result := make([]int, n)
	copy(result, arr)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if result[j] > result[j+1] {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}
	return result
}

func SelectionSort(arr []int) []int {
	n := len(arr)
	result := make([]int, n)
	copy(result, arr)
	for i := 0; i < n-1; i++ {
		minIdx := i
		for j := i + 1; j < n; j++ {
			if result[j] < result[minIdx] {
				minIdx = j
			}
		}
		result[i], result[minIdx] = result[minIdx], result[i]
	}
	return result
}

func InsertionSort(arr []int) []int {
	n := len(arr)
	result := make([]int, n)
	copy(result, arr)
	for i := 1; i < n; i++ {
		key := result[i]
		j := i - 1
		for j >= 0 && result[j] > key {
			result[j+1] = result[j]
			j--
		}
		result[j+1] = key
	}
	return result
}

func MergeSort(arr []int) []int {
	n := len(arr)
	if n <= 1 {
		result := make([]int, n)
		copy(result, arr)
		return result
	}
	mid := n / 2
	left := MergeSort(arr[:mid])
	right := MergeSort(arr[mid:])
	return merge(left, right)
}

func merge(left, right []int) []int {
	result := make([]int, 0, len(left)+len(right))
	i, j := 0, 0
	for i < len(left) && j < len(right) {
		if left[i] <= right[j] {
			result = append(result, left[i])
			i++
		} else {
			result = append(result, right[j])
			j++
		}
	}
	result = append(result, left[i:]...)
	result = append(result, right[j:]...)
	return result
}

func QuickSort(arr []int) []int {
	n := len(arr)
	result := make([]int, n)
	copy(result, arr)
	quickSortHelper(result, 0, n-1)
	return result
}

func quickSortHelper(arr []int, low, high int) {
	if low < high {
		pivot := partition(arr, low, high)
		quickSortHelper(arr, low, pivot-1)
		quickSortHelper(arr, pivot+1, high)
	}
}

func partition(arr []int, low, high int) int {
	pivot := arr[high]
	i := low - 1
	for j := low; j < high; j++ {
		if arr[j] < pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}

func HeapSort(arr []int) []int {
	n := len(arr)
	result := make([]int, n)
	copy(result, arr)
	for i := n/2 - 1; i >= 0; i-- {
		heapify(result, n, i)
	}
	for i := n - 1; i > 0; i-- {
		result[0], result[i] = result[i], result[0]
		heapify(result, i, 0)
	}
	return result
}

func heapify(arr []int, n, i int) {
	largest := i
	left := 2*i + 1
	right := 2*i + 2
	if left < n && arr[left] > arr[largest] {
		largest = left
	}
	if right < n && arr[right] > arr[largest] {
		largest = right
	}
	if largest != i {
		arr[i], arr[largest] = arr[largest], arr[i]
		heapify(arr, n, largest)
	}
}

func ShellSort(arr []int) []int {
	n := len(arr)
	result := make([]int, n)
	copy(result, arr)
	for gap := n / 2; gap > 0; gap /= 2 {
		for i := gap; i < n; i++ {
			temp := result[i]
			j := i
			for j >= gap && result[j-gap] > temp {
				result[j] = result[j-gap]
				j -= gap
			}
			result[j] = temp
		}
	}
	return result
}

func CombSort(arr []int) []int {
	n := len(arr)
	result := make([]int, n)
	copy(result, arr)
	gap := n
	shrink := 1.3
	sorted := false
	for !sorted {
		gap = int(float64(gap) / shrink)
		if gap <= 1 {
			gap = 1
			sorted = true
		}
		for i := 0; i+gap < n; i++ {
			if result[i] > result[i+gap] {
				result[i], result[i+gap] = result[i+gap], result[i]
				sorted = false
			}
		}
	}
	return result
}

func CocktailSort(arr []int) []int {
	n := len(arr)
	result := make([]int, n)
	copy(result, arr)
	swapped := true
	start := 0
	end := n - 1
	for swapped {
		swapped = false
		for i := start; i < end; i++ {
			if result[i] > result[i+1] {
				result[i], result[i+1] = result[i+1], result[i]
				swapped = true
			}
		}
		if !swapped {
			break
		}
		swapped = false
		end--
		for i := end - 1; i >= start; i-- {
			if result[i] > result[i+1] {
				result[i], result[i+1] = result[i+1], result[i]
				swapped = true
			}
		}
		start++
	}
	return result
}

func GnomeSort(arr []int) []int {
	n := len(arr)
	result := make([]int, n)
	copy(result, arr)
	i := 0
	for i < n {
		if i == 0 || result[i] >= result[i-1] {
			i++
		} else {
			result[i], result[i-1] = result[i-1], result[i]
			i--
		}
	}
	return result
}

func StdSort(arr []int) []int {
	result := make([]int, len(arr))
	copy(result, arr)
	sort.Ints(result)
	return result
}

func RadixSort(arr []int) []int {
	n := len(arr)
	if n == 0 {
		return []int{}
	}
	result := make([]int, n)
	copy(result, arr)
	maxVal := result[0]
	for _, v := range result {
		if v > maxVal {
			maxVal = v
		}
	}
	for exp := 1; maxVal/exp > 0; exp *= 10 {
		countingSort(result, exp)
	}
	return result
}

func countingSort(arr []int, exp int) {
	n := len(arr)
	output := make([]int, n)
	count := make([]int, 10)
	for i := range n {
		index := (arr[i] / exp) % 10
		count[index]++
	}
	for i := 1; i < 10; i++ {
		count[i] += count[i-1]
	}
	for i := n - 1; i >= 0; i-- {
		index := (arr[i] / exp) % 10
		output[count[index]-1] = arr[i]
		count[index]--
	}
	copy(arr, output)
}
