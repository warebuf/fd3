package compressState

import (
	"fmt"
	"strconv"
	"strings"
)

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

		fmt.Println(i, "result", result.String(), "s[i]", string(s[i]), "prev", string(prev), "curr", string(curr), "count", count)

		if s[i] == ';' {
			result.WriteString(strconv.Itoa(count))
			result.WriteByte(prev)
			result.WriteByte(';')
			prev = byte('$')
			count = 0
			fmt.Println("c", string(prev), count)
		} else if s[i] == ',' {
			fmt.Println("b", string(prev), count)
		} else if prev == '$' {
			prev = curr
			count = 1
			fmt.Println("a", string(prev), count)
		} else if prev == curr {
			count += 1
			fmt.Println("d", string(prev), count)
		} else {
			result.WriteString(strconv.Itoa(count))
			result.WriteByte(prev)
			prev = curr
			count = 1
			fmt.Println("e", string(prev), count)
		}

	}

	// flush last run
	result.WriteString(strconv.Itoa(count))
	result.WriteByte(prev)

	return result.String()
}
