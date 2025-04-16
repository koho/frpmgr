package util

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"unicode/utf8"
)

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

// RandToken generates a random hex value.
func RandToken(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
