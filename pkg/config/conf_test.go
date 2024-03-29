package config

import (
	"os"
	"testing"
	"time"
)

func init() {
	if err := os.MkdirAll("testdata", 0750); err != nil {
		panic(err)
	}
	if err := os.Chdir("testdata"); err != nil {
		panic(err)
	}
}

func TestExpiry(t *testing.T) {
	if err := os.WriteFile("example.ini", []byte("test"), 0666); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		input    AutoDelete
		expected time.Duration
	}{
		{input: AutoDelete{DeleteMethod: "relative", DeleteAfterDays: 5}, expected: 5 * time.Hour * 24},
		{input: AutoDelete{DeleteMethod: "absolute", DeleteAfterDate: time.Now().AddDate(0, 0, 3)}, expected: 3 * time.Hour * 24},
	}
	for i, test := range tests {
		output, err := Expiry("example.ini", test.input)
		if err != nil {
			t.Error(err)
			continue
		}
		if (test.expected - output).Abs() > 3*time.Second {
			t.Errorf("Test %d: expected: %v, got: %v", i, test.expected, output)
		}
	}
}
