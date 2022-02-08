package config

import (
	"github.com/koho/frpmgr/utils"
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type AuthInfo struct {
	AuthMethod        string `ini:"authentication_method,omitempty"`
	Token             string `ini:"token,omitempty"`
	OIDCClientId      string `ini:"oidc_client_id,omitempty"`
	OIDCClientSecret  string `ini:"oidc_client_secret,omitempty"`
	OIDCAudience      string `ini:"oidc_audience,omitempty"`
	OIDCTokenEndpoint string `ini:"oidc_token_endpoint_url,omitempty"`
}

type Common struct {
	AuthInfo             `ini:"common"`
	ServerAddress        string            `ini:"server_addr"`
	ServerPort           string            `ini:"server_port"`
	ConnectServerLocalIP string            `ini:"connect_server_local_ip,omitempty"`
	HTTPProxy            string            `ini:"http_proxy,omitempty"`
	LogFile              string            `ini:"log_file,omitempty"`
	LogLevel             string            `ini:"log_level,omitempty"`
	LogMaxDays           uint              `ini:"log_max_days,omitempty"`
	AdminAddr            string            `ini:"admin_addr,omitempty"`
	AdminPort            string            `ini:"admin_port,omitempty"`
	AdminUser            string            `ini:"admin_user,omitempty"`
	AdminPwd             string            `ini:"admin_pwd,omitempty"`
	PoolCount            uint              `ini:"pool_count,omitempty"`
	DNSServer            string            `ini:"dns_server,omitempty"`
	TcpMux               bool              `ini:"tcp_mux,omitempty"`
	Protocol             string            `ini:"protocol,omitempty"`
	TLSEnable            bool              `ini:"tls_enable,omitempty"`
	LoginFailExit        bool              `ini:"login_fail_exit"`
	User                 string            `ini:"user,omitempty"`
	ManualStart          bool              `ini:"manual_start,omitempty"`
	Custom               map[string]string `ini:"-"`
}

type Section struct {
	Name           string            `ini:"-"`
	Type           string            `ini:"type,omitempty"`
	LocalIP        string            `ini:"local_ip,omitempty"`
	LocalPort      string            `ini:"local_port,omitempty"`
	RemotePort     string            `ini:"remote_port,omitempty"`
	Role           string            `ini:"role,omitempty"`
	SK             string            `ini:"sk,omitempty"`
	ServerName     string            `ini:"server_name,omitempty"`
	BindAddr       string            `ini:"bind_addr,omitempty"`
	BindPort       string            `ini:"bind_port,omitempty"`
	UseEncryption  bool              `ini:"use_encryption,omitempty"`
	UseCompression bool              `ini:"use_compression,omitempty"`
	Custom         map[string]string `ini:"-"`
}

type Config struct {
	Name   string `ini:"-"`
	Path   string
	Status ServiceState
	Common
	Items []*Section
}

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
	c.Common.Custom = make(map[string]string)
	for _, key := range common.Keys() {
		if utils.GetFieldName(key.Name(), "ini", Common{}) == "" {
			c.Common.Custom[key.Name()] = key.String()
		}
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
	for k, v := range c.Common.Custom {
		common.Key(k).SetValue(v)
	}
	for _, item := range c.Items {
		s, err := cfg.NewSection(item.Name)
		if err != nil {
			return err
		}
		s.ReflectFrom(&item)
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
