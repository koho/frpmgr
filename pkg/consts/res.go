package consts

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// Icons
const (
	IconLogo         = 11
	IconOpen         = 22
	IconRefresh      = 23
	IconCopy         = 24
	IconCopyActive   = 25
	IconNewConf      = 149
	IconCreate       = 205
	IconImport       = 132
	IconDelete       = 131
	IconExport       = -174
	IconQuickAdd     = 111
	IconEdit         = 269
	IconRemote       = 20
	IconComputer     = 94
	IconWeb          = 13
	IconNewVersion1  = -1028
	IconNewVersion2  = 1
	IconStateRunning = 101
	IconStateStopped = 21
	IconStateWorking = 238
)

// Colors
var (
	ColorBlue  = walk.RGB(11, 53, 137)
	ColorGreen = walk.RGB(0, 100, 0)
)

// Text
var (
	TextRegular = Font{Family: "微软雅黑", PointSize: 9}
	TextMiddle  = Font{Family: "微软雅黑", PointSize: 12}
	TextLarge   = Font{Family: "微软雅黑", PointSize: 16}
)
