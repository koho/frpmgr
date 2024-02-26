package ui

import (
	"github.com/lxn/walk"

	"github.com/koho/frpmgr/pkg/consts"
)

var cachedIconsForWidthAndId = make(map[widthAndId]*walk.Icon)

func loadIcon(id consts.Icon, size int) (icon *walk.Icon) {
	icon = cachedIconsForWidthAndId[widthAndId{size, id}]
	if icon != nil {
		return
	}
	var err error
	if id.Dll == "" {
		icon, err = walk.NewIconFromResourceIdWithSize(id.Index, walk.Size{Width: size, Height: size})
	} else {
		icon, err = walk.NewIconFromSysDLLWithSize(id.Dll, id.Index, size)
	}
	if err == nil {
		cachedIconsForWidthAndId[widthAndId{size, id}] = icon
	}
	return
}

type widthAndId struct {
	width int
	icon  consts.Icon
}

type widthAndState struct {
	width int
	state consts.ServiceState
}

var cachedIconsForWidthAndState = make(map[widthAndState]*walk.Icon)

func iconForState(state consts.ServiceState, size int) (icon *walk.Icon) {
	icon = cachedIconsForWidthAndState[widthAndState{size, state}]
	if icon != nil {
		return
	}
	switch state {
	case consts.StateStarted:
		icon = loadIcon(consts.IconStateRunning, size)
	case consts.StateStopped, consts.StateUnknown:
		icon = loadIcon(consts.IconStateStopped, size)
	default:
		icon = loadIcon(consts.IconStateWorking, size)
	}
	cachedIconsForWidthAndState[widthAndState{size, state}] = icon
	return
}

func loadLogoIcon(size int) *walk.Icon {
	return loadIcon(consts.IconLogo, size)
}

func loadShieldIcon(size int) (icon *walk.Icon) {
	icon = loadIcon(consts.IconNewVersion1, size)
	if icon == nil {
		icon = loadIcon(consts.IconNewVersion2, size)
	}
	return
}

func drawCopyIcon(canvas *walk.Canvas, color walk.Color) error {
	brush, err := walk.NewSolidColorBrush(color)
	if err != nil {
		return err
	}
	defer brush.Dispose()

	pen, err := walk.NewGeometricPen(walk.PenSolid|walk.PenInsideFrame|walk.PenCapSquare|walk.PenJoinMiter, 2, brush)
	if err != nil {
		return err
	}
	defer pen.Dispose()

	if err = canvas.DrawRectangle(pen, walk.Rectangle{X: 5, Y: 5, Width: 8, Height: 9}); err != nil {
		return err
	}
	// Outer line: (2, 2) -> (10, 2)
	if err = canvas.DrawLine(pen, walk.Point{X: 3, Y: 3}, walk.Point{X: 9, Y: 3}); err != nil {
		return err
	}
	// Outer line: (2, 2) -> (2, 11)
	if err = canvas.DrawLine(pen, walk.Point{X: 3, Y: 3}, walk.Point{X: 3, Y: 10}); err != nil {
		return err
	}
	return nil
}
