package compressState

import (
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
	result.WriteString(strconv.Itoa(count))
	result.WriteByte(prev)

	return result.String()
}
