package config

import (
	"encoding/json"
	"os"

	"github.com/koho/frpmgr/pkg/consts"
)

const (
	DefaultAppFile = "app.json"
	LangFile       = "lang.config"
)

type App struct {
	Lang        string       `json:"lang,omitempty"`
	Password    string       `json:"password,omitempty"`
	CheckUpdate bool         `json:"checkUpdate"`
	Defaults    DefaultValue `json:"defaults"`
	Sort        []string     `json:"sort,omitempty"`
	Position    []int32      `json:"position,omitempty"`
}

type DefaultValue struct {
	Protocol             string `json:"protocol,omitempty"`
	User                 string `json:"user,omitempty"`
	LogLevel             string `json:"logLevel"`
	LogMaxDays           int64  `json:"logMaxDays"`
	DNSServer            string `json:"dnsServer,omitempty"`
	NatHoleSTUNServer    string `json:"natHoleStunServer,omitempty"`
	ConnectServerLocalIP string `json:"connectServerLocalIP,omitempty"`
	TCPMux               bool   `json:"tcpMux"`
	TLSEnable            bool   `json:"tls"`
	ManualStart          bool   `json:"manualStart,omitempty"`
	LegacyFormat         bool   `json:"legacyFormat,omitempty"`
}

func (dv *DefaultValue) AsClientConfig() ClientCommon {
	return ClientCommon{
		ServerPort:                consts.DefaultServerPort,
		Protocol:                  dv.Protocol,
		User:                      dv.User,
		LogLevel:                  dv.LogLevel,
		LogMaxDays:                dv.LogMaxDays,
		DNSServer:                 dv.DNSServer,
		NatHoleSTUNServer:         dv.NatHoleSTUNServer,
		ConnectServerLocalIP:      dv.ConnectServerLocalIP,
		TCPMux:                    dv.TCPMux,
		TLSEnable:                 dv.TLSEnable,
		ManualStart:               dv.ManualStart,
		LegacyFormat:              dv.LegacyFormat,
		DisableCustomTLSFirstByte: true,
	}
}

func UnmarshalAppConf(path string, dst *App) (lang *string, err error) {
	b, err := os.ReadFile(LangFile)
	if err == nil {
		s := string(b)
		lang = &s
	}
	b, err = os.ReadFile(path)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, dst)
	return
}

func (conf *App) Save(path string) error {
	b, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0666)
}
