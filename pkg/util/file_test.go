package util

import (
	"os"
	"reflect"
	"testing"
)

func TestSplitExt(t *testing.T) {
	tests := []struct {
		input        string
		expectedName string
		expectedExt  string
	}{
		{input: "C:\\test\\a.ini", expectedName: "a", expectedExt: ".ini"},
		{input: "b.exe", expectedName: "b", expectedExt: ".exe"},
		{input: "c", expectedName: "c", expectedExt: ""},
		{input: "", expectedName: "", expectedExt: ""},
	}
	for i, test := range tests {
		name, ext := SplitExt(test.input)
		if name != test.expectedName {
			t.Errorf("Test %d: expected: %v, got: %v", i, test.expectedName, name)
		}
		if ext != test.expectedExt {
			t.Errorf("Test %d: expected: %v, got: %v", i, test.expectedExt, ext)
		}
	}
}

func TestFindLogFiles(t *testing.T) {
	tests := []struct {
		create        []string
		expectedFiles []string
		expectedDates []string
	}{
		{
			create:        []string{"example.log", "example.2023-03-20.log", "example.2023-03-21.log", "example.2023-03-21T01.log"},
			expectedFiles: []string{"example.log", "example.2023-03-20.log", "example.2023-03-21.log"},
			expectedDates: []string{"", "2023-03-20", "2023-03-21"},
		},
	}
	if err := os.MkdirAll("testdata", 0750); err != nil {
		t.Fatal(err)
	}
	os.Chdir("testdata")
	for i, test := range tests {
		for _, f := range test.create {
			os.WriteFile(f, []byte("test"), 0666)
		}
		logs, dates, err := FindLogFiles(test.create[0])
		if err != nil {
			t.Error(err)
			continue
		}
		if !reflect.DeepEqual(logs, test.expectedFiles) {
			t.Errorf("Test %d: expected: %v, got: %v", i, test.expectedFiles, logs)
		}
		if !reflect.DeepEqual(dates, test.expectedDates) {
			t.Errorf("Test %d: expected: %v, got: %v", i, test.expectedDates, dates)
		}
	}
}

func TestAddFileSuffix(t *testing.T) {
	tests := []struct {
		input    string
		suffix   string
		expected string
	}{
		{input: "C:\\test\\a.ini", suffix: "_1", expected: "a_1.ini"},
		{input: "b.exe", suffix: "_2", expected: "b_2.exe"},
		{input: "c", suffix: "_3", expected: "c_3"},
		{input: "", suffix: "_4", expected: "_4"},
		{input: "", suffix: "", expected: ""},
	}
	for i, test := range tests {
		output := AddFileSuffix(test.input, test.suffix)
		if output != test.expected {
			t.Errorf("Test %d: expected: %v, got: %v", i, test.expected, output)
		}
	}
}
