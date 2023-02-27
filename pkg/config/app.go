package config

import (
	"gopkg.in/ini.v1"
)

const DefaultAppFile = "app.config"

type App struct {
	Password string       `ini:"password,omitempty"`
	Defaults ClientCommon `ini:"defaults"`
}

func UnmarshalAppConfFromIni(source any, dst *App) error {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
		AllowBooleanKeys:    true,
	}, source)
	if err != nil {
		return err
	}
	if err = cfg.MapTo(dst); err != nil {
		return err
	}
	return nil
}

func (conf *App) Save(path string) error {
	cfg := ini.Empty()
	if err := cfg.ReflectFrom(conf); err != nil {
		return err
	}
	return cfg.SaveTo(path)
}
