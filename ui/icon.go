package ui

import (
	"github.com/lxn/walk"

	"github.com/koho/frpmgr/pkg/consts"
)

var cachedSystemIconsForWidthAndDllIdx = make(map[widthDllIdx]*walk.Icon)

func loadSysIcon(dll string, index int32, size int) (icon *walk.Icon) {
	icon = cachedSystemIconsForWidthAndDllIdx[widthDllIdx{size, index, dll}]
	if icon != nil {
		return
	}
	var err error
	icon, err = walk.NewIconFromSysDLLWithSize(dll, int(index), size)
	if err == nil {
		cachedSystemIconsForWidthAndDllIdx[widthDllIdx{size, index, dll}] = icon
	}
	return
}

type widthDllIdx struct {
	width int
	idx   int32
	dll   string
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
		icon = loadSysIcon("imageres", consts.IconStateRunning, size)
	case consts.StateStopped, consts.StateUnknown:
		icon = loadResourceIcon(consts.IconStateStopped, size)
	default:
		icon = loadSysIcon("shell32", consts.IconStateWorking, size)
	}
	cachedIconsForWidthAndState[widthAndState{size, state}] = icon
	return
}

func loadLogoIcon(size int) *walk.Icon {
	return loadResourceIcon(consts.IconLogo, size)
}

var cachedResourceIcons = make(map[widthDllIdx]*walk.Icon)

func loadResourceIcon(id int, size int) (icon *walk.Icon) {
	icon = cachedResourceIcons[widthDllIdx{width: size, idx: int32(id)}]
	if icon != nil {
		return
	}
	var err error
	icon, err = walk.NewIconFromResourceIdWithSize(id, walk.Size{Width: size, Height: size})
	if err == nil {
		cachedResourceIcons[widthDllIdx{width: size, idx: int32(id)}] = icon
	}
	return
}
