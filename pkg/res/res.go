package res

import (
	"fmt"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"github.com/samber/lo"
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

type Icon struct {
	Dll   string
	Index int
}

// Icons
var (
	IconLogo         = Icon{Index: 7}
	IconOpen         = Icon{"imageres", -5339}
	IconRandom       = Icon{"imageres", -1024}
	IconSysCopy      = Icon{"shell32", -243}
	IconNewConf      = Icon{"shell32", -258}
	IconCreate       = Icon{"shell32", -319}
	IconFileImport   = Icon{"shell32", -241}
	IconURLImport    = Icon{"imageres", -184}
	IconClipboard    = Icon{"shell32", -16763}
	IconDelete       = Icon{"shell32", -240}
	IconExport       = Icon{"imageres", -174}
	IconQuickAdd     = Icon{"shell32", -16769}
	IconEdit         = Icon{"shell32", -16775}
	IconEnable       = Icon{"shell32", -16810}
	IconDisable      = Icon{"imageres", -1027}
	IconEditDialog   = Icon{"imageres", -114}
	IconRemote       = Icon{"imageres", -25}
	IconSSH          = Icon{"imageres", -5372}
	IconVNC          = Icon{"imageres", -110}
	IconWeb          = Icon{"shell32", -14}
	IconDns          = Icon{"imageres", -145}
	IconFtp          = Icon{"imageres", -143}
	IconHttpFile     = Icon{"imageres", -73}
	IconHttpProxy    = Icon{"imageres", -120}
	IconOpenPort     = Icon{"shell32", -244}
	IconLock         = Icon{"shell32", -48}
	IconNewVersion1  = Icon{"imageres", -1028}
	IconNewVersion2  = Icon{"imageres", 1}
	IconUpdate       = Icon{"shell32", -47}
	IconStateRunning = Icon{"imageres", -106}
	IconStateStopped = Icon{Index: 21}
	IconStateWorking = Icon{"shell32", -16739}
	IconDefaults     = Icon{"imageres", -165}
	IconKey          = Icon{"imageres", -5360}
	IconLanguage     = Icon{"imageres", -94}
	IconNat          = Icon{"firewallcontrolpanel", 0}
	IconExperiment   = Icon{"imageres", -111}
	IconFile         = Icon{"shell32", -152}
	IconInfo         = Icon{"imageres", -81}
	IconArrowUp      = Icon{"shell32", -16817}
	IconMove         = Icon{"imageres", -5313}
)

// Colors
var (
	ColorBlue      = walk.RGB(0, 38, 247)
	ColorDarkBlue  = walk.RGB(11, 53, 137)
	ColorLightBlue = walk.RGB(49, 94, 251)
	ColorGray      = walk.RGB(109, 109, 109)
	ColorDarkGray  = walk.RGB(85, 85, 85)
	ColorGrayBG    = walk.Color(win.GetSysColor(win.COLOR_BTNFACE))
)

// Text
var (
	TextRegular Font
	TextMedium  Font
	TextLarge   Font
)

func init() {
	var defaultFontFamily string
	versionInfo := windows.RtlGetVersion()
	if versionInfo.MajorVersion > 6 || (versionInfo.MajorVersion == 6 && versionInfo.MinorVersion > 1) {
		// > Windows 7
		defaultFontFamily = "Microsoft YaHei UI"
	} else {
		// <= Windows 7
		// Microsoft YaHei UI is not included in Windows 7
		// Fallback to Microsoft YaHei instead
		defaultFontFamily = "Microsoft YaHei"
		// Fallback icons
		IconKey.Index = -82
		IconOpen = Icon{"shell32", -46}
		IconSSH.Index = -100
		IconEnable.Index = -253
	}
	TextRegular = Font{Family: defaultFontFamily, PointSize: 9}
	TextMedium = Font{Family: defaultFontFamily, PointSize: 10}
	TextLarge = Font{Family: defaultFontFamily, PointSize: 16}
}

var (
	SupportedConfigFormats = []string{".ini", ".toml", ".json", ".yml", ".yaml"}
	cfgPatterns            = lo.Map(append([]string{".zip"}, SupportedConfigFormats...), func(item string, index int) string {
		return "*" + item
	})
)

// Filters
var (
	FilterAllFiles = i18n.Sprintf("All Files") + " (*.*)|*.*"
	FilterConfig   = i18n.Sprintf("Configuration Files") + fmt.Sprintf(" (%s)|%s|", strings.Join(cfgPatterns, ", "), strings.Join(cfgPatterns, ";"))
	FilterZip      = i18n.Sprintf("Configuration Files") + " (*.zip)|*.zip"
	FilterCert     = i18n.Sprintf("Certificate Files") + " (*.crt, *.cer)|*.crt;*.cer|"
	FilterKey      = i18n.Sprintf("Key Files") + " (*.key)|*.key|"
	FilterLog      = i18n.Sprintf("Log Files") + " (*.log, *.txt)|*.log;*.txt|"
)

// Validators
var (
	ValidateNonEmpty = validators.Regexp{Pattern: "[^\\s]+"}
)

// Dialogs
const (
	DialogValidate = "VALDLG"
	DialogTitle    = 2000
	DialogStatic1  = 2001
	DialogStatic2  = 2002
	DialogEdit     = 2003
	DialogIcon     = 2004
)
