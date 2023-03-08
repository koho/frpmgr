package validators

import (
	"github.com/lxn/walk"
)

type RegexpValidator struct {
	*walk.RegexpValidator
}

func NewRegexpValidator(pattern string) (*RegexpValidator, error) {
	re, err := walk.NewRegexpValidator(pattern)
	if err != nil {
		return nil, err
	}

	return &RegexpValidator{re}, nil
}

func (rv *RegexpValidator) Validate(v interface{}) error {
	err := rv.RegexpValidator.Validate(v)
	if str, ok := v.(string); ok && str == "" && err != nil {
		return silentErr
	}
	return err
}

type Regexp struct {
	Pattern string
}

func (re Regexp) Create() (walk.Validator, error) {
	return NewRegexpValidator(re.Pattern)
}
