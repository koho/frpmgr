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
	IconSysCopy      = 134
	IconNewConf      = 149
	IconCreate       = 205
	IconImport       = 132
	IconDelete       = 131
	IconExport       = -174
	IconQuickAdd     = 111
	IconEdit         = 269
	IconEnable       = 27
	IconDisable      = 28
	IconEditDialog   = -114
	IconRemote       = 20
	IconSSH          = 26
	IconVNC          = 105
	IconWeb          = 13
	IconDns          = 139
	IconFtp          = 137
	IconHttpFile     = 69
	IconHttpProxy    = 114
	IconSocks5       = 146
	IconVpn          = 47
	IconNewVersion1  = -1028
	IconNewVersion2  = 1
	IconUpdate       = -47
	IconStateRunning = 101
	IconStateStopped = 21
	IconStateWorking = 238
)

// Colors
var (
	ColorBlue = walk.RGB(11, 53, 137)
	ColorGray = walk.RGB(109, 109, 109)
)

// Text
var (
	TextRegular = Font{Family: "微软雅黑", PointSize: 9}
	TextMedium  = Font{Family: "微软雅黑", PointSize: 10}
	TextLarge   = Font{Family: "微软雅黑", PointSize: 16}
)

// Filters
const (
	FilterAllFiles = "所有文件 (*.*)|*.*"
	FilterConfig   = "配置文件 (*.zip, *.ini)|*.zip;*.ini|"
	FilterZip      = "配置文件 (*.zip)|*.zip"
	FilterCert     = "证书文件 (*.crt, *.cer)|*.crt;*.cer|"
	FilterKey      = "密钥文件 (*.key)|*.key|"
)

// Validators
var (
	ValidateNonEmpty       = Regexp{Pattern: ".+"}
	ValidateRequireInteger = Regexp{Pattern: "^\\d+$"}
	ValidateInteger        = Regexp{Pattern: "^\\d*$"}
)
