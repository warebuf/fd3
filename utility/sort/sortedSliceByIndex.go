package sort

import "sort"

func SortedIndexes(arr []int) []int {
	n := len(arr)
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}

	sort.Slice(idx, func(i, j int) bool {
		return arr[idx[i]] < arr[idx[j]]
	})

	return idx
}
