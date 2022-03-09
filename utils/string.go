package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

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

func Partition(s string, sep string) (string, string, string) {
	parts := strings.SplitN(s, sep, 2)
	if len(parts) == 1 {
		return parts[0], "", ""
	}
	return parts[0], sep, parts[1]
}

func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func Map2String(m map[string]string) string {
	sb := strings.Builder{}
	for k, v := range m {
		sb.WriteString(fmt.Sprintf("%s = %s\r\n", k, v))
	}
	return sb.String()
}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
