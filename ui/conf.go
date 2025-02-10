package ui

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

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
	sync.Mutex
	// Name of the config
	Name        string
	DisplayName string
	// Path of the config file
	Path string
	// State of service
	State consts.ServiceState
	// Install indicates whether a service is installed
	Install bool
	// Data is ClientConfig or ServerConfig
	Data config.Config
}

// PathOfConf returns the file path of a config with given base file name
func PathOfConf(base string) string {
	return filepath.Join("profiles", base)
}

func NewConf(path string, data config.Config) *Conf {
	conf := &Conf{Path: path, Data: data}
	baseName, _ := util.SplitExt(path)
	conf.Name = baseName
	return conf
}

func (conf *Conf) FileNameWithoutExt() string {
	if conf.Path == "" {
		return ""
	}
	return strings.TrimSuffix(filepath.Base(conf.Path), filepath.Ext(conf.Path))
}

// Delete config will remove service, logs, config file in disk/mem
func (conf *Conf) Delete() (bool, error) {
	// Delete service
	running := conf.State == consts.StateStarted
	if err := services.UninstallService(conf.Name, true); err != nil && running {
		return false, err
	}
	// Delete logs
	if logs, _, err := util.FindLogFiles(conf.Data.GetLogFile()); err == nil {
		util.DeleteFiles(logs)
	}
	// Delete config file
	if err := os.Remove(conf.Path); err != nil {
		return false, err
	}
	// Delete mem config
	return deleteConf(conf), nil
}

// Save config to the disk. The config will be completed before saving
func (conf *Conf) Save() error {
	conf.Data.Complete(false)
	conf.Path = PathOfConf(conf.Name + ".conf")
	return conf.Data.Save(conf.Path)
}

var (
	appConf = config.App{Defaults: config.DefaultValue{
		LogLevel:   consts.LogLevelInfo,
		LogMaxDays: consts.DefaultLogMaxDays,
		TCPMux:     true,
		TLSEnable:  true,
	}}
	// The config list contains all the loaded configs
	confList  []*Conf
	confMutex sync.Mutex
	confDB    *walk.DataBinder
)

func loadAllConfs() error {
	_ = config.UnmarshalAppConf(config.DefaultAppFile, &appConf)
	// Find all config files in `profiles` directory
	files, err := filepath.Glob(PathOfConf("*.conf"))
	if err != nil {
		return err
	}
	confList = make([]*Conf, 0)
	for _, f := range files {
		c := NewConf(f, nil)
		if conf, err := config.UnmarshalClientConf(f); err == nil {
			c.Data = conf
			if conf.DeleteAfterDays > 0 {
				if t, err := config.Expiry(f, conf.AutoDelete); err == nil && t <= 0 {
					c.Delete()
					continue
				}
			}
			confList = append(confList, c)
		}
	}
	slices.SortStableFunc(confList, func(a, b *Conf) int {
		i := slices.Index(appConf.Sort, a.FileNameWithoutExt())
		j := slices.Index(appConf.Sort, b.FileNameWithoutExt())
		if i < 0 && j >= 0 {
			return 1
		} else if j < 0 && i >= 0 {
			return -1
		}
		return i - j
	})
	return nil
}

// Make a copy of config list
func getConfListSafe() []*Conf {
	confMutex.Lock()
	t := append([]*Conf(nil), confList...)
	confMutex.Unlock()
	return t
}

// Add a new config to the mem config list
func addConf(conf *Conf) {
	confMutex.Lock()
	defer confMutex.Unlock()
	confList = append(confList, conf)
	setConfOrder()
	saveAppConfig()
}

// Remove a config from the mem config list
func deleteConf(conf *Conf) bool {
	confMutex.Lock()
	defer confMutex.Unlock()
	for i := range confList {
		if confList[i] == conf {
			confList = append(confList[:i], confList[i+1:]...)
			setConfOrder()
			saveAppConfig()
			return true
		}
	}
	return false
}

// Check whether a config exists with the given name
func hasConf(name string) bool {
	return slices.ContainsFunc(confList, func(e *Conf) bool { return e.Name == name })
}

// ConfBinder is the view model of the current selected config
type ConfBinder struct {
	// Current selected config
	Current *Conf
	// Selected indicates whether there's a selected config
	Selected bool
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
			ds.Selected = ds.Current != nil
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

func newDefaultClientConfig() *config.ClientConfig {
	return &config.ClientConfig{
		ClientCommon: appConf.Defaults.AsClientConfig(),
	}
}

func saveAppConfig() error {
	return appConf.Save(config.DefaultAppFile)
}

func setConfOrder() {
	appConf.Sort = lo.Map(confList, func(item *Conf, index int) string {
		return item.FileNameWithoutExt()
	})
}
