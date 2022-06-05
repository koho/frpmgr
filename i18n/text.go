package i18n

//go:generate gotext -srclang=en-US update -out=catalog.go -lang=en-US,zh-CN,zh-TW,ja-JP,ko-KR,es-ES ../cmd/frpmgr

import (
	"bufio"
	"golang.org/x/sys/windows"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"os"
	"path/filepath"
	"strings"
)

var printer *message.Printer

func init() {
	if preferredLang := langInConfig(); preferredLang != "" {
		printer = message.NewPrinter(language.Make(preferredLang))
	} else {
		printer = message.NewPrinter(lang())
	}
}

// langInConfig returns the UI language code in config file
func langInConfig() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	langFile, err := os.Open(filepath.Join(filepath.Dir(exePath), "lang.config"))
	if err != nil {
		return ""
	}
	defer langFile.Close()

	scanner := bufio.NewScanner(langFile)
	for scanner.Scan() {
		if text := strings.TrimSpace(scanner.Text()); text != "" && !strings.HasPrefix(text, "#") {
			return text
		}
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
