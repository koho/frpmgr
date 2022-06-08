package consts

import (
	"github.com/koho/frpmgr/i18n"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.org/x/sys/windows"
)

// Links
const (
	ProjectURL    = "https://github.com/koho/frpmgr"
	FRPProjectURL = "https://github.com/fatedier/frp"
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
	IconFileImport   = 132
	IconClipboard    = 260
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
	defaultFontFamily = func() string {
		versionInfo := windows.RtlGetVersion()
		// Microsoft YaHei UI is not included in Windows 7
		// Fallback to Microsoft YaHei instead
		if versionInfo.MajorVersion > 6 || (versionInfo.MajorVersion == 6 && versionInfo.MinorVersion > 1) {
			// > Windows 7
			return "Microsoft YaHei UI"
		} else {
			// <= Windows 7
			return "Microsoft YaHei"
		}
	}()
	TextRegular = Font{Family: defaultFontFamily, PointSize: 9}
	TextMedium  = Font{Family: defaultFontFamily, PointSize: 10}
	TextLarge   = Font{Family: defaultFontFamily, PointSize: 16}
)

// Filters
var (
	FilterAllFiles = i18n.Sprintf("All Files") + " (*.*)|*.*"
	FilterConfig   = i18n.Sprintf("Configuration Files") + " (*.zip, *.ini)|*.zip;*.ini|"
	FilterZip      = i18n.Sprintf("Configuration Files") + " (*.zip)|*.zip"
	FilterCert     = i18n.Sprintf("Certificate Files") + " (*.crt, *.cer)|*.crt;*.cer|"
	FilterKey      = i18n.Sprintf("Key Files") + " (*.key)|*.key|"
	FilterLog      = i18n.Sprintf("Log Files") + " (*.log, *.txt)|*.log;*.txt|"
)

// Validators
var (
	ValidateNonEmpty       = Regexp{Pattern: ".+"}
	ValidateRequireInteger = Regexp{Pattern: "^\\d+$"}
	ValidateInteger        = Regexp{Pattern: "^\\d*$"}
)
