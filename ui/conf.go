package ui

import (
	"errors"
	"os"
	"path/filepath"
	"slices"

	"github.com/lxn/walk"
	"github.com/samber/lo"

	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/util"
	"github.com/koho/frpmgr/services"
)

// The flag controls the running state of service.
type runFlag int

const (
	runFlagAuto runFlag = iota
	runFlagForceStart
	runFlagReload
)

// Conf contains all data of a config
type Conf struct {
	// Path of the config file
	Path string
	// State of service
	State consts.ConfigState
	// Data is ClientConfig or ServerConfig
	Data config.Config
}

// PathOfConf returns the file path of a config with given base file name
func PathOfConf(base string) string {
	return filepath.Join("profiles", base)
}

func NewConf(path string, data config.Config) *Conf {
	if path == "" {
		filename, err := util.RandToken(16)
		if err != nil {
			panic(err)
		}
		path = PathOfConf(filename + ".conf")
	}
	return &Conf{
		Path:  path,
		State: consts.ConfigStateStopped,
		Data:  data,
	}
}

func (conf *Conf) Name() string {
	return conf.Data.Name()
}

// Delete config will remove service, logs, config file in disk
func (conf *Conf) Delete() error {
	// Delete service
	running := conf.State == consts.ConfigStateStarted
	if err := services.UninstallService(conf.Path, true); err != nil && running {
		return err
	}
	// Delete logs
	if logs, _, err := util.FindLogFiles(conf.Data.GetLogFile()); err == nil {
		util.DeleteFiles(logs)
	}
	// Delete config file
	if err := os.Remove(conf.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// Save config to the disk. The config will be completed before saving
func (conf *Conf) Save() error {
	logPath, err := filepath.Abs(filepath.Join("logs", util.FileNameWithoutExt(conf.Path)+".log"))
	if err != nil {
		return err
	}
	conf.Data.Complete(false)
	conf.Data.SetLogFile(filepath.ToSlash(logPath))
	return conf.Data.Save(conf.Path)
}

var (
	appConf = config.App{Defaults: config.DefaultValue{
		LogLevel:   consts.LogLevelInfo,
		LogMaxDays: consts.DefaultLogMaxDays,
		TCPMux:     true,
		TLSEnable:  true,
	}}
	confDB *walk.DataBinder
)

func loadAllConfs() ([]*Conf, error) {
	_ = config.UnmarshalAppConf(config.DefaultAppFile, &appConf)
	// Find all config files in `profiles` directory
	files, err := filepath.Glob(PathOfConf("*.conf"))
	if err != nil {
		return nil, err
	}
	cfgList := make([]*Conf, 0)
	for _, f := range files {
		if conf, err := config.UnmarshalClientConf(f); err == nil {
			c := NewConf(f, conf)
			if c.Name() == "" {
				conf.ClientCommon.Name = util.FileNameWithoutExt(f)
			}
			cfgList = append(cfgList, c)
		}
	}
	slices.SortStableFunc(cfgList, func(a, b *Conf) int {
		i := slices.Index(appConf.Sort, util.FileNameWithoutExt(a.Path))
		j := slices.Index(appConf.Sort, util.FileNameWithoutExt(b.Path))
		if i < 0 && j >= 0 {
			return 1
		} else if j < 0 && i >= 0 {
			return -1
		}
		return i - j
	})
	return cfgList, nil
}

// ConfBinder is the view model of configs
type ConfBinder struct {
	// Current selected config
	Current *Conf
	// List of configs
	List func() []*Conf
	// Set Config state
	SetState func(conf *Conf, state consts.ConfigState) bool
	// Commit will save the given config and try to reload service
	Commit func(conf *Conf, flag runFlag)
}

// getCurrentConf returns the current selected config
func getCurrentConf() *Conf {
	if confDB != nil {
		if ds, ok := confDB.DataSource().(*ConfBinder); ok {
			return ds.Current
		}
	}
	return nil
}

// setCurrentConf set the current selected config, the views will get notified
func setCurrentConf(conf *Conf) {
	if confDB != nil {
		if ds, ok := confDB.DataSource().(*ConfBinder); ok {
			ds.Current = conf
			confDB.Reset()
		}
	}
}

// commitConf will save the given config and try to reload service
func commitConf(conf *Conf, flag runFlag) {
	if confDB != nil {
		if ds, ok := confDB.DataSource().(*ConfBinder); ok {
			ds.Commit(conf, flag)
		}
	}
}

// getConfList returns a list of all configs.
func getConfList() []*Conf {
	if confDB != nil {
		if ds, ok := confDB.DataSource().(*ConfBinder); ok {
			return ds.List()
		}
	}
	return nil
}

func setConfState(conf *Conf, state consts.ConfigState) bool {
	if confDB != nil {
		if ds, ok := confDB.DataSource().(*ConfBinder); ok {
			return ds.SetState(conf, state)
		}
	}
	return false
}

func newDefaultClientConfig() *config.ClientConfig {
	return &config.ClientConfig{
		ClientCommon: appConf.Defaults.AsClientConfig(),
	}
}

func saveAppConfig() error {
	return appConf.Save(config.DefaultAppFile)
}

func setConfOrder(cfgList []*Conf) {
	appConf.Sort = lo.Map(cfgList, func(item *Conf, index int) string {
		return util.FileNameWithoutExt(item.Path)
	})
	saveAppConfig()
}
