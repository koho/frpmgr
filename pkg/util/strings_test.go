package util

import (
	"reflect"
	"strings"
	"testing"
)

func slice2Map(elems []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, x := range elems {
		m[x] = struct{}{}
	}
	return m
}

func TestString2Map(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]string
	}{
		{input: "a= 1\nb  =2\nc  = 3", expected: map[string]string{"a": "1", "b": "2", "c": "3"}},
		{input: "a=4", expected: map[string]string{"a": "4"}},
		{input: "a = 5", expected: map[string]string{"a": "5"}},
		{input: "", expected: map[string]string{}},
	}
	for i, test := range tests {
		output := String2Map(test.input)
		if !reflect.DeepEqual(output, test.expected) {
			t.Errorf("Test %d: expected: %v, got: %v", i, test.expected, output)
		}
	}
}

func TestMap2String(t *testing.T) {
	tests := []struct {
		input    map[string]string
		expected string
	}{
		{input: map[string]string{"a": "1 st", "b": "2", "c": "3"}, expected: "a = 1 st\nb = 2\nc = 3\n"},
		{input: map[string]string{}, expected: ""},
	}
	for i, test := range tests {
		output := Map2String(test.input)
		outLines := strings.Split(output, "\r\n")
		expLines := strings.Split(test.expected, "\n")
		if !reflect.DeepEqual(slice2Map(outLines), slice2Map(expLines)) {
			t.Errorf("Test %d: expected: %v, got: %v", i, expLines, outLines)
		}
	}
}

func TestPartition(t *testing.T) {
	tests := []struct {
		input   string
		sep     string
		l, s, r string
	}{
		{input: "a-b", sep: "-", l: "a", s: "-", r: "b"},
		{input: "ab", sep: "-", l: "ab", s: "", r: ""},
		{input: "ab-", sep: "-", l: "ab", s: "-", r: ""},
		{input: "ab", sep: "+", l: "ab", s: "", r: ""},
		{input: "", sep: "-", l: "", s: "", r: ""},
	}
	for i, test := range tests {
		left, sep, right := Partition(test.input, test.sep)
		if left != test.l || sep != test.s || right != test.r {
			t.Errorf("Test %d: expected: %v, got: %v", i, []string{test.l, test.s, test.r}, []string{left, sep, right})
		}
	}
}

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
