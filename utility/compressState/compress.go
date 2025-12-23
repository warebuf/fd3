package compressState

import (
	"strconv"
	"strings"
)

func IntToAlphaNumeric(nums []int) string {

	mapChar := map[int]byte{
		1: 'h',
		2: 'a',
		3: 'c',
	}

	// create an array of length n (max num in mapChar)
	counter := make([]int, 3+1)
	for i := 0; i < len(nums); i++ {
		counter[nums[i]] += 1
	}

	var result strings.Builder
	for i := 0; i < len(counter); i++ {
		if counter[i] > 0 {
			result.WriteString(strconv.Itoa(counter[i]))
			result.WriteByte(mapChar[i])
		}
	}

	return result.String()
}

func CompressPattern(s string) string {
	if len(s) == 0 {
		return ""
	}

	mapChar := map[byte]byte{
		'1': 'h',
		'2': 'a',
		'3': 'c',
	}

	var result strings.Builder

	prev := byte('$')
	count := 0
	for i := 0; i < len(s); i++ {

		curr := mapChar[s[i]]

		if prev == '$' && s[i] == ';' {
			result.WriteByte(';')
		} else if prev == '$' {
			prev = curr
			count = 1
		} else if s[i] == ',' {
		} else if s[i] == ';' {
			result.WriteString(strconv.Itoa(count))
			result.WriteByte(prev)
			result.WriteByte(';')
			prev = byte('$')
			count = 0
		} else if prev == curr {
			count += 1
		} else {
			result.WriteString(strconv.Itoa(count))
			result.WriteByte(prev)
			prev = curr
			count = 1
		}

	}

	// flush last run
	if prev != '$' {
		result.WriteString(strconv.Itoa(count))
		result.WriteByte(prev)
	}

	return result.String()
}
