package consts

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/validators"
)

// Links
const (
	ProjectURL      = "https://github.com/koho/frpmgr"
	FRPProjectURL   = "https://github.com/fatedier/frp"
	UpdateURL       = "https://api.github.com/repos/koho/frpmgr/releases/latest"
	ShareLinkScheme = "frp://"
)

// Icons
const (
	IconLogo         = 7
	IconOpen         = 22
	IconRefresh      = 23
	IconCopy         = 24
	IconCopyActive   = 25
	IconSysCopy      = 134
	IconNewConf      = 149
	IconCreate       = 205
	IconFileImport   = 132
	IconURLImport    = 175
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
	IconOpenPort     = 135
	IconVpn          = 47
	IconNewVersion1  = -1028
	IconNewVersion2  = 1
	IconUpdate       = -47
	IconStateRunning = 101
	IconStateStopped = 21
	IconStateWorking = 238
	IconDefaults     = 156
	IconKey          = 29
	IconLanguage     = 89
	IconNat          = 0
)

// Colors
var (
	ColorBlue     = walk.RGB(0, 38, 247)
	ColorDarkBlue = walk.RGB(11, 53, 137)
	ColorGray     = walk.RGB(109, 109, 109)
	ColorGrayBG   = walk.Color(win.GetSysColor(win.COLOR_BTNFACE))
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
	ValidateNonEmpty       = validators.Regexp{Pattern: "[^\\s]+"}
	ValidateRequireInteger = validators.Regexp{Pattern: "^\\d+$"}
	ValidateInteger        = validators.Regexp{Pattern: "^\\d*$"}
	ValidatePortRange      = []Validator{ValidateRequireInteger, validators.Range{Min: 0, Max: 65535}}
)

// Dialogs
const (
	DialogValidate = "VALDLG"
	DialogTitle    = 2000
	DialogStatic1  = 2001
	DialogStatic2  = 2002
	DialogEdit     = 2003
)
