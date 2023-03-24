package validators

import (
	"errors"
	"testing"
)

func TestRegexp(t *testing.T) {
	r, err := Regexp{Pattern: "^\\d+$"}.Create()
	if err != nil {
		t.Fatal(err)
	}
	if err = r.Validate(""); !errors.Is(err, silentErr) {
		t.Errorf("Expected: %v, got: %v", silentErr, err)
	}
	tests := []struct {
		input     string
		shouldErr bool
	}{
		{input: "123", shouldErr: false},
		{input: "a1", shouldErr: true},
		{input: "1.1", shouldErr: true},
		{input: " 1", shouldErr: true},
		{input: "1a", shouldErr: true},
	}
	for i, test := range tests {
		err = r.Validate(test.input)
		if (test.shouldErr && err == nil) || (!test.shouldErr && err != nil) {
			t.Errorf("Test %d: expected: %v, got: %v", i, test.shouldErr, err)
		}
	}
}
