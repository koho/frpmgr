package config

import (
	"frpmgr/utils"
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type AuthInfo struct {
	AuthMethod        string `ini:"authentication_method"`
	Token             string `ini:"token"`
	OIDCClientId      string `ini:"oidc_client_id"`
	OIDCClientSecret  string `ini:"oidc_client_secret"`
	OIDCAudience      string `ini:"oidc_audience"`
	OIDCTokenEndpoint string `ini:"oidc_token_endpoint_url"`
}

type Common struct {
	ServerAddress string `ini:"server_addr"`
	ServerPort    string `ini:"server_port"`
	LogFile       string `ini:"log_file"`
	LogLevel      string `ini:"log_level"`
	LogMaxDays    uint   `ini:"log_max_days"`
	AuthInfo      `ini:"common"`
	AdminAddr     string `ini:"admin_addr"`
	AdminPort     string `ini:"admin_port"`
	AdminUser     string `ini:"admin_user"`
	AdminPwd      string `ini:"admin_pwd"`
	PoolCount     uint   `ini:"pool_count"`
	DNSServer     string `ini:"dns_server"`
	TcpMux        bool   `ini:"tcp_mux"`
	Protocol      string `ini:"protocol"`
	TLSEnable     bool   `ini:"tls_enable"`
	LoginFailExit bool   `ini:"login_fail_exit"`
	ManualStart   bool   `ini:"manual_start"`
}

type Section struct {
	Name           string            `ini:"-"`
	Type           string            `ini:"type"`
	LocalIP        string            `ini:"local_ip"`
	LocalPort      string            `ini:"local_port"`
	RemotePort     string            `ini:"remote_port"`
	Role           string            `ini:"role"`
	SK             string            `ini:"sk"`
	ServerName     string            `ini:"server_name"`
	BindAddr       string            `ini:"bind_addr"`
	BindPort       string            `ini:"bind_port"`
	UseEncryption  bool              `ini:"use_encryption"`
	UseCompression bool              `ini:"use_compression"`
	Custom         map[string]string `ini:"-"`
}

type Config struct {
	Name   string `ini:"-"`
	Path   string
	Status ServiceState
	Common
	Items []*Section
}

var notOmitEmpty = []string{"login_fail_exit"}
var Configurations []*Config
var ConfMutex sync.Mutex
var StatusChan = make(chan bool)

func LoadConfig() ([]*Config, error) {
	files, err := filepath.Glob("*.ini")
	if err != nil {
		return nil, err
	}
	confList := make([]*Config, 0)
	for _, f := range files {
		c := new(Config)
		if err = c.Load(f); err != nil {
			continue
		}
		confList = append(confList, c)
	}
	return confList, nil
}

func GetConfigNames() []string {
	names := make([]string, 0)
	for _, c := range Configurations {
		names = append(names, c.Name)
	}
	return names
}

func (c *Config) GetSectionNames() []string {
	names := make([]string, 0)
	for _, sect := range c.Items {
		names = append(names, sect.Name)
	}
	return names
}

func (c *Config) Load(source string) error {
	c.Name = NameFromPath(source)
	c.Path = source
	c.Status = StateStopped
	cfg, err := ini.Load(source)
	if err != nil {
		return err
	}
	common, err := cfg.GetSection("common")
	if err != nil {
		return err
	}
	err = common.MapTo(&c.Common)
	if err != nil {
		return err
	}
	c.Items = make([]*Section, 0)
	for _, section := range cfg.Sections() {
		if section.Name() == "common" || section.Name() == "DEFAULT" {
			continue
		}
		s := Section{Name: section.Name()}
		section.MapTo(&s)
		s.Custom = make(map[string]string)
		for _, key := range section.Keys() {
			if utils.GetFieldName(key.Name(), "ini", Section{}) == "" {
				s.Custom[key.Name()] = key.String()
			}
		}
		c.Items = append(c.Items, &s)
	}
	return nil
}

func (c *Config) SaveTo(path string) error {
	cfg := ini.Empty()
	common, err := cfg.NewSection("common")
	if err != nil {
		return err
	}
	common.ReflectFrom(&c.Common)
	for _, k := range common.Keys() {
		if _, found := utils.Find(notOmitEmpty, k.Name()); found {
			continue
		}
		if k.Value() == "" || k.Value() == "0" || k.Value() == "false" {
			common.DeleteKey(k.Name())
		}
	}
	for _, item := range c.Items {
		s, err := cfg.NewSection(item.Name)
		if err != nil {
			return err
		}
		s.ReflectFrom(&item)
		for _, sk := range s.Keys() {
			if sk.Value() == "" || sk.Value() == "0" || sk.Value() == "false" {
				s.DeleteKey(sk.Name())
			}
		}
		for k, v := range item.Custom {
			s.Key(k).SetValue(v)
		}
	}
	for _, sect := range cfg.Sections() {
		if sect.Name() == "common" || sect.Name() == "DEFAULT" {
			continue
		}
		if _, found := utils.Find(c.GetSectionNames(), sect.Name()); !found {
			cfg.DeleteSection(sect.Name())
		}
	}
	return cfg.SaveTo(path)
}

func (c *Config) Save() error {
	return c.SaveTo(c.Name + ".ini")
}

func (c *Config) Delete() error {
	return os.Remove(c.Name + ".ini")
}

func NameFromPath(confPath string) string {
	name := filepath.Base(confPath)
	return strings.TrimSuffix(name, filepath.Ext(name))
}

func PathFromName(name string) (string, error) {
	return filepath.Abs(name + ".ini")
}
