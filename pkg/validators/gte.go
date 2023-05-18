package validators

import (
	"github.com/lxn/walk"

	"github.com/koho/frpmgr/i18n"
)

type GTEValidator struct {
	Value **walk.LineEdit
}

func (g *GTEValidator) Validate(v interface{}) error {
	text := v.(string)
	f, err := walk.ParseFloat(text)
	if err != nil {
		return nanErr
	}
	val := (*g.Value).Text()
	if val == "" {
		return silentErr
	}
	t, err := walk.ParseFloat(val)
	if err != nil {
		return nanErr
	}
	if f >= t {
		return nil
	}
	return walk.NewValidationError(i18n.Sprintf("Number out of allowed range"),
		i18n.Sprintf("Please enter a number greater than %d.", int(t)))
}

// GTE checks whether the input value is greater than or equal to the target value.
type GTE struct {
	Value **walk.LineEdit
}

func (g GTE) Create() (walk.Validator, error) {
	return &GTEValidator{g.Value}, nil
}
