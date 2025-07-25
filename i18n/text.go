package i18n

//go:generate go run golang.org/x/text/cmd/gotext -srclang=en-US update -out=catalog.go -lang=en-US,zh-CN,zh-TW,ja-JP,ko-KR,es-ES ../cmd/frpmgr

import (
	"encoding/json"
	"os"

	"golang.org/x/sys/windows"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/koho/frpmgr/pkg/config"
)

var (
	printer  *message.Printer
	useLang  language.Tag
	IDToName = map[string]string{
		"zh-CN": "简体中文",
		"zh-TW": "繁體中文",
		"en-US": "English",
		"ja-JP": "日本語",
		"ko-KR": "한국어",
		"es-ES": "Español",
	}
)

func init() {
	if preferredLang := langInConfig(); preferredLang != "" {
		useLang = language.Make(preferredLang)
	} else {
		useLang = lang()
	}
	printer = message.NewPrinter(useLang)
}

// GetLanguage returns the current display language code.
func GetLanguage() string {
	return useLang.String()
}

// langInConfig returns the UI language code in config file
func langInConfig() string {
	b, err := os.ReadFile(config.LangFile)
	if err == nil {
		id := string(b)
		if _, ok := IDToName[id]; ok {
			return id
		}
	}
	b, err = os.ReadFile(config.DefaultAppFile)
	if err != nil {
		return ""
	}
	var s struct {
		Lang string `json:"lang"`
	}
	if err = json.Unmarshal(b, &s); err != nil {
		return ""
	}
	if _, ok := IDToName[s.Lang]; ok {
		return s.Lang
	}
	return ""
}

// lang returns the user preferred UI language.
func lang() (tag language.Tag) {
	tag = language.English
	languages, err := windows.GetUserPreferredUILanguages(windows.MUI_LANGUAGE_NAME)
	if err != nil {
		return
	}
	if match := message.MatchLanguage(languages...); !match.IsRoot() {
		tag = match
	}
	return
}

// Sprintf is just a wrapper function of message printer.
func Sprintf(key message.Reference, a ...interface{}) string {
	return printer.Sprintf(key, a...)
}

// SprintfColon adds a colon at the tail of a string.
func SprintfColon(key message.Reference, a ...interface{}) string {
	return Sprintf(key, a...) + ":"
}

// SprintfEllipsis adds an ellipsis at the tail of a string.
func SprintfEllipsis(key message.Reference, a ...interface{}) string {
	return Sprintf(key, a...) + "..."
}

// SprintfLSpace adds a space at the start of a string.
func SprintfLSpace(key message.Reference, a ...interface{}) string {
	return " " + Sprintf(key, a...)
}

// SprintfRSpace adds a space at the end of a string.
func SprintfRSpace(key message.Reference, a ...interface{}) string {
	return Sprintf(key, a...) + " "
}
