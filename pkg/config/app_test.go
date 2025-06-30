package config

import (
	"os"
	"reflect"
	"testing"
)

func TestUnmarshalAppConfFromIni(t *testing.T) {
	input := `{
	"password": "abcde",
	"defaults": {
		"logLevel": "info",
		"logMaxDays": 5,
		"protocol": "kcp",
		"user": "user",
		"tcpMux": true,
		"manualStart": true,
		"legacyFormat": true
	}
}
	`
	if err := os.WriteFile(DefaultAppFile, []byte(input), 0666); err != nil {
		t.Fatal(err)
	}
	expected := App{
		Password: "abcde",
		Defaults: DefaultValue{
			LogLevel:     "info",
			LogMaxDays:   5,
			Protocol:     "kcp",
			User:         "user",
			TCPMux:       true,
			ManualStart:  true,
			LegacyFormat: true,
		},
	}
	expectedLang := "en-US"
	if err := os.WriteFile(LangFile, []byte(expectedLang), 0666); err != nil {
		t.Fatal(err)
	}
	var actual App
	lang, err := UnmarshalAppConf(DefaultAppFile, &actual)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected: %v, got: %v", expected, actual)
	}
	if lang == nil || *lang != expectedLang {
		t.Errorf("Expected: %v, got: %v", expectedLang, lang)
	}
}
