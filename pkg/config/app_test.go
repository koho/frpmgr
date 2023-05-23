package config

import (
	"reflect"
	"testing"
)

func TestUnmarshalAppConfFromIni(t *testing.T) {
	input := `
		password = abcde

		[defaults]
		server_port = 7000
		log_level = info
		log_max_days = 5
		protocol = kcp
		login_fail_exit = false
		user = user
		tcp_mux = true
		frpmgr_manual_start = true
		frpmgr_delete_after_days = 1
	`
	expected := App{
		Password: "abcde",
		Defaults: ClientCommon{
			ServerPort:    "7000",
			LogLevel:      "info",
			LogMaxDays:    5,
			Protocol:      "kcp",
			LoginFailExit: false,
			User:          "user",
			TCPMux:        true,
			ManualStart:   true,
			AutoDelete: AutoDelete{
				DeleteAfterDays: 1,
			},
		},
	}
	var actual App
	if err := UnmarshalAppConfFromIni([]byte(input), &actual); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected: %v, got: %v", expected, actual)
	}
}
