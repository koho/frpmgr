package validators

import "github.com/lxn/walk"

// TextEqual compares to the text of a LineEdit.
type TextEqual struct {
	Target **walk.LineEdit
}

func (t TextEqual) Create() (walk.Validator, error) {
	return &TextEqual{t.Target}, nil
}

func (t *TextEqual) Validate(v interface{}) error {
	text := v.(string)
	if (*t.Target).Text() == text {
		return nil
	}
	return walk.NewValidationError("Text mismatch", "")
}
