package utils

import (
	"strings"
)

func SnakeToCamel(s string) string {
	var buf []byte
	strBytes := []byte(s)
	for i, c := range strBytes {
		switch {
		case 'a' <= c && c <= 'z':
			if i == 0 || strBytes[i-1] == '-' || strBytes[i-1] == '_' {
				buf = append(buf, []byte(strings.ToUpper(string(c)))...)
			} else {
				buf = append(buf, c)
			}
		case c == '-' || c == '_':
			continue
		}
	}
	return string(buf)
}
