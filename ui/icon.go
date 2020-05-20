package ui

import "github.com/lxn/walk"

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
