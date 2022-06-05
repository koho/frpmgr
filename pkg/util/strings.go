package util

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// String2Map splits the string by new line character. Each line is partitioned by equal sign.
// The left side is the key while the right side is the value.
func String2Map(s string) map[string]string {
	m := make(map[string]string)
	for _, line := range strings.Split(s, "\n") {
		key, sep, value := Partition(line, "=")
		if sep == "" {
			continue
		}
		m[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return m
}

// Map2String turns the map into string. The key value pairs are separated by equal sign.
// Each pair is separated by new line character.
func Map2String(m map[string]string) string {
	sb := strings.Builder{}
	for k, v := range m {
		sb.WriteString(fmt.Sprintf("%s = %s\r\n", k, v))
	}
	return sb.String()
}

// Partition returns the left side of the separator, the separator and the right side of the separator.
func Partition(s string, sep string) (string, string, string) {
	parts := strings.SplitN(s, sep, 2)
	if len(parts) == 1 {
		return parts[0], "", ""
	}
	return parts[0], sep, parts[1]
}

// GetOrElse returns the given string if it's non-empty, or returns the default string.
func GetOrElse(s string, def string) string {
	if strings.TrimSpace(s) != "" {
		return s
	}
	return def
}

// RuneSizeInString returns a slice of each character's size in the given string
func RuneSizeInString(s string) []int {
	sizes := make([]int, 0)
	for len(s) > 0 {
		_, size := utf8.DecodeRuneInString(s)
		sizes = append(sizes, size)
		s = s[size:]
	}
	return sizes
}
