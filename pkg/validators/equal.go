package validators

import "github.com/lxn/walk"

type TextEqualValidator struct {
	Target **walk.LineEdit
}

func (t *TextEqualValidator) Validate(v interface{}) error {
	text := v.(string)
	if (*t.Target).Text() == text {
		return nil
	}
	return walk.NewValidationError("Text mismatch", "")
}

// TextEqual checks whether the input text is equal to the target value.
type TextEqual struct {
	Target **walk.LineEdit
}

func (t TextEqual) Create() (walk.Validator, error) {
	return &TextEqualValidator{t.Target}, nil
}
