package util

import (
	"reflect"
	"testing"
)

func TestGetOrElse(t *testing.T) {
	tests := []struct {
		input    string
		def      string
		expected string
	}{
		{input: "abc", def: "def", expected: "abc"},
		{input: "", def: "def", expected: "def"},
		{input: " ", def: "def", expected: "def"},
	}
	for i, test := range tests {
		output := GetOrElse(test.input, test.def)
		if output != test.expected {
			t.Errorf("Test %d: expected: %v, got: %v", i, test.expected, output)
		}
	}
}

func TestRuneSizeInString(t *testing.T) {
	str := "Hello, 世界"
	expected := []int{1, 1, 1, 1, 1, 1, 1, 3, 3}
	output := RuneSizeInString(str)
	if !reflect.DeepEqual(output, expected) {
		t.Errorf("Expected: %v, got: %v", expected, output)
	}
}
