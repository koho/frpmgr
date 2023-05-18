package validators

import (
	"github.com/lxn/walk"

	"github.com/koho/frpmgr/i18n"
)

type PasswordValidator struct {
	Password **walk.LineEdit
}

func (p *PasswordValidator) Validate(v interface{}) error {
	text := v.(string)
	if text == "" {
		return silentErr
	}
	if (*p.Password).Text() == text {
		return nil
	}
	return walk.NewValidationError(i18n.Sprintf("Password mismatch"), i18n.Sprintf("Please check and try again."))
}

// ConfirmPassword checks whether the input text is equal to the password field.
type ConfirmPassword struct {
	Password **walk.LineEdit
}

func (c ConfirmPassword) Create() (walk.Validator, error) {
	return &PasswordValidator{c.Password}, nil
}
