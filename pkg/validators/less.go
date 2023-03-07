package validators

import "github.com/lxn/walk"

type LessThanValidator struct {
	Target **walk.LineEdit
}

func (l *LessThanValidator) Validate(v interface{}) error {
	text := v.(string)
	f, err := walk.ParseFloat(text)
	if err != nil {
		return walk.NewValidationError("Not a number", "")
	}
	t, err := walk.ParseFloat((*l.Target).Text())
	if err != nil {
		return walk.NewValidationError("Not a number", "")
	}
	if f <= t {
		return nil
	}
	return walk.NewValidationError("Number out of allowed range", "")
}

// LessThan checks whether the input value is less than or equal to the target value.
type LessThan struct {
	Target **walk.LineEdit
}

func (l LessThan) Create() (walk.Validator, error) {
	return &LessThanValidator{l.Target}, nil
}
