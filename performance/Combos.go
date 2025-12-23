package performance

import (
	"fmt"
	"sort"

	"github.com/warebuf/fd3/engine"
)

func CheckMaxSlice(arr []int, max int) bool {
	for i := 0; i < len(arr); i++ {
		if arr[i] != max {
			return false
		}
	}
	return true
}

func GenerateCombos(N int, models []engine.Model) [][]int {

	combos := [][]int{}

	// init a slice of 1s (skip 0 lol)
	arr := make([]int, N)
	for i := 0; i < len(arr); i++ {
		arr[i] = 1
	}

	max := len(models) - 1

	// increment the last digit and propagate up
	for {

		// save the shallow copy
		snapshot := make([]int, len(arr))
		copy(snapshot, arr)
		combos = append(combos, snapshot)

		// checks if we are done all possible combinations
		if CheckMaxSlice(arr, max) {
			break
		}

		// increment the last digit and propagate up the carry forward
		for j := len(arr) - 1; j >= 0; j-- {

			// increment the digit
			arr[j] += 1

			// carry forward or stop
			if arr[j] > max {
				arr[j] = 1
			} else {
				break
			}

		}

	}

	return combos
}

func GenerateTeams(combos [][]int) [][2][]int {
	teams := [][2][]int{}
	for i := 0; i < len(combos); i++ {
		for j := 1; j < len(combos[i]); j++ {
			a := combos[i][j:]
			b := combos[i][:j]
			teams = append(teams, [2][]int{b, a})
		}
	}
	return teams
}

func getKey(a []int, b []int) string {
	return fmt.Sprintf("%v;%v", a, b)
}

func cloneInts(x []int) []int {
	y := make([]int, len(x))
	copy(y, x)
	return y
}

// sort and remove dups
func RemoveDups(inputs [][2][]int) [][2][]int {

	cleaned := [][2][]int{}
	seen := map[string]bool{}

	for i := 0; i < len(inputs); i++ {
		a := cloneInts(inputs[i][0])
		b := cloneInts(inputs[i][1])
		sort.Ints(a)
		sort.Ints(b)
		key := getKey(a, b)
		if _, ok := seen[key]; !ok {
			seen[key] = true
			cleaned = append(cleaned, [2][]int{a, b})
		}
	}

	return cleaned
}
