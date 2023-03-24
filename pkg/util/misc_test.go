package util

import "testing"

type tagTest struct {
	Tag  string
	Name string `t1:"true" t2:"true"`
	Age  int    `t2:"true"`
}

func TestPruneByTag(t *testing.T) {
	tests := []struct {
		input    tagTest
		expected tagTest
	}{
		{input: tagTest{Tag: "t1", Name: "John", Age: 34}, expected: tagTest{Name: "John"}},
		{input: tagTest{Tag: "t2", Name: "Ben", Age: 20}, expected: tagTest{Name: "Ben", Age: 20}},
		{input: tagTest{Name: "Mary", Age: 50}, expected: tagTest{}},
	}
	for i, test := range tests {
		output, err := PruneByTag(test.input, "true", test.input.Tag)
		if err != nil {
			t.Fatalf("Test %d: expected no error but found one for input %v, got: %v", i, test.input, err)
		}
		if output != test.expected {
			t.Errorf("Test %d: expected: %v, got: %v", i, test.expected, output)
		}
	}
}

func TestGetFieldNameByTag(t *testing.T) {
	tests := []struct {
		tag      string
		value    string
		expected string
	}{
		{tag: "t1", value: "true", expected: "Name"},
		{tag: "t2", value: "true", expected: "Name"},
		{tag: "t1", value: "false", expected: ""},
		{tag: "t3", value: "true", expected: ""},
	}
	for i, test := range tests {
		output := GetFieldNameByTag(tagTest{}, test.tag, test.value)
		if output != test.expected {
			t.Errorf("Test %d: expected: %v, got: %v", i, test.expected, output)
		}
	}
}
