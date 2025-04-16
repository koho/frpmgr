package validators

import (
	"errors"

	"github.com/lxn/walk"
)

var silentErr = errors.New("")

type ToolTipErrorPresenter struct {
	*walk.ToolTipErrorPresenter
}

func NewToolTipErrorPresenter() (*ToolTipErrorPresenter, error) {
	p, err := walk.NewToolTipErrorPresenter()
	if err != nil {
		return nil, err
	}
	return &ToolTipErrorPresenter{p}, nil
}

func (ttep *ToolTipErrorPresenter) PresentError(err error, widget walk.Widget) {
	if errors.Is(err, silentErr) {
		ttep.ToolTipErrorPresenter.PresentError(nil, widget)
	} else {
		ttep.ToolTipErrorPresenter.PresentError(err, widget)
	}
}

// SilentToolTipErrorPresenter hides the tooltip when the input value is empty.
type SilentToolTipErrorPresenter struct {
}

func (SilentToolTipErrorPresenter) Create() (walk.ErrorPresenter, error) {
	return NewToolTipErrorPresenter()
}
