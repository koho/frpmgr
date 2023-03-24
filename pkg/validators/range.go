package validators

import (
	"github.com/lxn/walk"
)

type RangeValidator struct {
	*walk.RangeValidator
}

func NewRangeValidator(min, max float64) (*RangeValidator, error) {
	validator, err := walk.NewRangeValidator(min, max)
	if err != nil {
		return nil, err
	}

	return &RangeValidator{validator}, nil
}

func (rv *RangeValidator) Validate(v interface{}) error {
	var value float64
	switch v := v.(type) {
	case string:
		f, err := walk.ParseFloat(v)
		if err != nil {
			return nanErr
		}
		value = f
	case float64:
		value = v
	default:
		panic("Unsupported type")
	}
	return rv.RangeValidator.Validate(value)
}

// Range checks whether the input value is between Min and Max value.
// Supported widgets: NumberEdit, LineEdit.
type Range struct {
	Min float64
	Max float64
}

func (r Range) Create() (walk.Validator, error) {
	return NewRangeValidator(r.Min, r.Max)
}
