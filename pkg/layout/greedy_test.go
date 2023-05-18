package layout

import (
	"testing"

	"github.com/lxn/walk"
)

func TestNewGreedyLayoutItem(t *testing.T) {
	tests := []struct {
		input                walk.Orientation
		expected, unexpected []walk.LayoutFlags
	}{
		{input: walk.Horizontal, expected: []walk.LayoutFlags{walk.GreedyHorz}, unexpected: []walk.LayoutFlags{walk.GreedyVert}},
		{input: walk.Vertical, expected: []walk.LayoutFlags{walk.GreedyVert}, unexpected: []walk.LayoutFlags{walk.GreedyHorz}},
		{input: walk.NoOrientation, expected: []walk.LayoutFlags{walk.GreedyHorz, walk.GreedyVert}, unexpected: nil},
	}
	for i, test := range tests {
		flags := NewGreedyLayoutItem(test.input).LayoutFlags()
		for _, f := range test.expected {
			if f&flags == 0 {
				t.Errorf("Test %d: expected: %v, got: %v", i, f, flags)
			}
		}
		for _, f := range test.unexpected {
			if f&flags > 0 {
				t.Errorf("Test %d: unexpected: %v, got: %v", i, f, flags)
			}
		}
	}
}
