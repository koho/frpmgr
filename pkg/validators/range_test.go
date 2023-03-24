package validators

import (
	"testing"
)

func TestRange(t *testing.T) {
	r, err := Range{Min: 1, Max: 200}.Create()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		input     any
		shouldErr bool
	}{
		{input: 0.0, shouldErr: true},
		{input: 1.1, shouldErr: false},
		{input: "1", shouldErr: false},
		{input: "1.1", shouldErr: false},
		{input: "201", shouldErr: true},
		{input: "a", shouldErr: true},
	}
	for i, test := range tests {
		err = r.Validate(test.input)
		if (test.shouldErr && err == nil) || (!test.shouldErr && err != nil) {
			t.Errorf("Test %d: expected: %v, got: %v", i, test.shouldErr, err)
		}
	}
}
