package config

import (
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/util"
	"github.com/thoas/go-funk"
	"gopkg.in/ini.v1"
)

type ClientAuth struct {
	AuthMethod        string `ini:"authentication_method,omitempty"`
	Token             string `ini:"token,omitempty" token:"true"`
	OIDCClientId      string `ini:"oidc_client_id,omitempty" oidc:"true"`
	OIDCClientSecret  string `ini:"oidc_client_secret,omitempty" oidc:"true"`
	OIDCAudience      string `ini:"oidc_audience,omitempty" oidc:"true"`
	OIDCTokenEndpoint string `ini:"oidc_token_endpoint_url,omitempty" oidc:"true"`
}

type ClientCommon struct {
	ClientAuth           `ini:",extends"`
	ServerAddress        string `ini:"server_addr"`
	ServerPort           string `ini:"server_port"`
	ConnectServerLocalIP string `ini:"connect_server_local_ip,omitempty"`
	HTTPProxy            string `ini:"http_proxy,omitempty"`
	LogFile              string `ini:"log_file,omitempty"`
	LogLevel             string `ini:"log_level,omitempty"`
	LogMaxDays           uint   `ini:"log_max_days,omitempty"`
	AdminAddr            string `ini:"admin_addr,omitempty"`
	AdminPort            string `ini:"admin_port,omitempty"`
	AdminUser            string `ini:"admin_user,omitempty"`
	AdminPwd             string `ini:"admin_pwd,omitempty"`
	PoolCount            uint   `ini:"pool_count,omitempty"`
	DNSServer            string `ini:"dns_server,omitempty"`
	Protocol             string `ini:"protocol,omitempty"`
	LoginFailExit        bool   `ini:"login_fail_exit"`
	User                 string `ini:"user,omitempty"`
	// Options for this project
	// ManualStart defines whether to start the config on system boot
	ManualStart bool `ini:"manual_start,omitempty"`
	// Custom collects all the unparsed options
	Custom map[string]string `ini:"-"`
}

// BaseProxyConf provides configuration info that is common to all types.
type BaseProxyConf struct {
	// Name is the name of this
	Name string `ini:"-"`
	// Type specifies the type of this  Valid values include tcp, udp,
	// xtcp, stcp, sudp, http, https, tcpmux. By default, this value is "tcp".
	Type string `ini:"type,omitempty"`

	// UseEncryption controls whether or not communication with the server will
	// be encrypted. Encryption is done using the tokens supplied in the server
	// and client configuration. By default, this value is false.
	UseEncryption bool `ini:"use_encryption,omitempty"`
	// UseCompression controls whether or not communication with the server
	// will be compressed. By default, this value is false.
	UseCompression bool `ini:"use_compression,omitempty"`
	// Group specifies which group the is a part of. The server will use
	// this information to load balance proxies in the same group. If the value
	// is "", this will not be in a group. By default, this value is "".
	Group string `ini:"group,omitempty"`
	// GroupKey specifies a group key, which should be the same among proxies
	// of the same group. By default, this value is "".
	GroupKey string `ini:"group_key,omitempty"`

	// ProxyProtocolVersion specifies which protocol version to use. Valid
	// values include "v1", "v2", and "". If the value is "", a protocol
	// version will be automatically selected. By default, this value is "".
	ProxyProtocolVersion string `ini:"proxy_protocol_version,omitempty"`

	// BandwidthLimit limit the bandwidth
	// 0 means no limit
	BandwidthLimit string `ini:"bandwidth_limit,omitempty"`

	// LocalIP specifies the IP address or host name.
	LocalIP string `ini:"local_ip,omitempty"`
	// LocalPort specifies the port.
	LocalPort string `ini:"local_port,omitempty"`

	// Plugin specifies what plugin should be used for ng. If this value
	// is set, the LocalIp and LocalPort values will be ignored. By default,
	// this value is "".
	Plugin string `ini:"plugin,omitempty"`
	// PluginParams specify parameters to be passed to the plugin, if one is
	// being used.
	PluginParams `ini:",extends"`
	// HealthCheckType specifies what protocol to use for health checking.
	HealthCheckType string `ini:"health_check_type,omitempty"` // tcp | http
	// Health checking parameters
	HealthCheckConf `ini:",extends"`
	// Custom collects all the unparsed options
	Custom map[string]string `ini:"-"`
}

type PluginParams struct {
	PluginLocalAddr         string `ini:"plugin_local_addr,omitempty" http2https:"true" https2https:"true" https2http:"true"`
	PluginCrtPath           string `ini:"plugin_crt_path,omitempty" https2https:"true" https2http:"true"`
	PluginKeyPath           string `ini:"plugin_key_path,omitempty" https2https:"true" https2http:"true"`
	PluginHostHeaderRewrite string `ini:"plugin_host_header_rewrite,omitempty" http2https:"true" https2https:"true" https2http:"true"`
	PluginHttpUser          string `ini:"plugin_http_user,omitempty" http_proxy:"true" static_file:"true"`
	PluginHttpPasswd        string `ini:"plugin_http_passwd,omitempty" http_proxy:"true" static_file:"true"`
	PluginUser              string `ini:"plugin_user,omitempty" socks5:"true"`
	PluginPasswd            string `ini:"plugin_passwd,omitempty" socks5:"true"`
	PluginLocalPath         string `ini:"plugin_local_path,omitempty" static_file:"true"`
	PluginStripPrefix       string `ini:"plugin_strip_prefix,omitempty" static_file:"true"`
	PluginUnixPath          string `ini:"plugin_unix_path,omitempty" unix_domain_socket:"true"`
}

// HealthCheckConf configures health checking. This can be useful for load
// balancing purposes to detect and remove proxies to failing services.
type HealthCheckConf struct {
	// HealthCheckTimeoutS specifies the number of seconds to wait for a health
	// check attempt to connect. If the timeout is reached, this counts as a
	// health check failure. By default, this value is 3.
	HealthCheckTimeoutS int `ini:"health_check_timeout_s,omitempty" tcp:"true" http:"true"`
	// HealthCheckMaxFailed specifies the number of allowed failures before the
	// is stopped. By default, this value is 1.
	HealthCheckMaxFailed int `ini:"health_check_max_failed,omitempty" tcp:"true" http:"true"`
	// HealthCheckIntervalS specifies the time in seconds between health
	// checks. By default, this value is 10.
	HealthCheckIntervalS int `ini:"health_check_interval_s,omitempty" tcp:"true" http:"true"`
	// HealthCheckURL specifies the address to send health checks to if the
	// health check type is "http".
	HealthCheckURL string `ini:"health_check_url,omitempty" http:"true"`
}

type Proxy struct {
	BaseProxyConf     `ini:",extends"`
	RemotePort        string `ini:"remote_port,omitempty" tcp:"true" udp:"true"`
	Role              string `ini:"role,omitempty" stcp:"true" xtcp:"true" sudp:"true" visitor:"true"`
	SK                string `ini:"sk,omitempty" stcp:"true" xtcp:"true" sudp:"true" visitor:"true"`
	ServerName        string `ini:"server_name,omitempty" visitor:"true"`
	BindAddr          string `ini:"bind_addr,omitempty" visitor:"true"`
	BindPort          string `ini:"bind_port,omitempty" visitor:"true"`
	CustomDomains     string `ini:"custom_domains,omitempty" http:"true" https:"true" tcpmux:"true"`
	SubDomain         string `ini:"subdomain,omitempty" http:"true" https:"true" tcpmux:"true"`
	Locations         string `ini:"locations,omitempty" http:"true"`
	HTTPUser          string `ini:"http_user,omitempty" http:"true"`
	HTTPPwd           string `ini:"http_pwd,omitempty" http:"true"`
	HostHeaderRewrite string `ini:"host_header_rewrite,omitempty" http:"true"`
	Multiplexer       string `ini:"multiplexer,omitempty" tcpmux:"true"`
}

func (p *Proxy) GetName() string {
	return p.Name
}

type ClientConfig struct {
	ClientCommon
	Proxies []*Proxy
}

func (conf *ClientConfig) AutoStart() bool {
	return !conf.ManualStart
}

func (conf *ClientConfig) GetLogFile() string {
	return conf.LogFile
}

func (conf *ClientConfig) Items() interface{} {
	return conf.Proxies
}

func (conf *ClientConfig) ItemAt(index int) interface{} {
	return conf.Proxies[index]
}

func (conf *ClientConfig) DeleteItem(index int) {
	conf.Proxies = append(conf.Proxies[:index], conf.Proxies[index+1:]...)
}

func (conf *ClientConfig) AddItem(item interface{}) bool {
	if proxy, ok := item.(*Proxy); ok {
		if !funk.Contains(conf.Proxies, func(p *Proxy) bool { return p.Name == proxy.Name }) {
			conf.Proxies = append(conf.Proxies, proxy)
			return true
		}
	}
	return false
}

func (conf *ClientConfig) Save(path string) error {
	cfg := ini.Empty()
	common, err := cfg.NewSection("common")
	if err != nil {
		return err
	}
	if err = common.ReflectFrom(&conf.ClientCommon); err != nil {
		return err
	}
	for k, v := range conf.ClientCommon.Custom {
		common.Key(k).SetValue(v)
	}
	for _, proxy := range conf.Proxies {
		p, err := cfg.NewSection(proxy.Name)
		if err != nil {
			return err
		}
		if err = p.ReflectFrom(&proxy); err != nil {
			return err
		}
		for k, v := range proxy.Custom {
			p.Key(k).SetValue(v)
		}
	}
	return cfg.SaveTo(path)
}

func (conf *ClientConfig) Complete() {
	// Common config
	authMethod := conf.AuthMethod
	if authMethod != "" {
		if auth, err := util.PruneByTag(conf.ClientAuth, "true", authMethod); err == nil {
			conf.ClientAuth = auth.(ClientAuth)
			conf.AuthMethod = authMethod
		}
	}
	// Proxies
	for _, proxy := range conf.Proxies {
		if funk.ContainsString([]string{consts.ProxyTypeXTCP, consts.ProxyTypeSTCP, consts.ProxyTypeSUDP}, proxy.Type) && proxy.Role == "visitor" {
			var base = proxy.BaseProxyConf
			// Visitor
			if p, err := util.PruneByTag(*proxy, "true", "visitor"); err == nil {
				*proxy = p.(Proxy)
			}
			proxy.BaseProxyConf = BaseProxyConf{Name: base.Name, Type: base.Type, UseEncryption: base.UseEncryption, UseCompression: base.UseCompression}
		} else {
			var base = proxy.BaseProxyConf
			// Plugins
			if base.Plugin != "" {
				if pluginParams, err := util.PruneByTag(base.PluginParams, "true", base.Plugin); err == nil {
					base.PluginParams = pluginParams.(PluginParams)
				}
			}
			// Health Check
			if base.HealthCheckType != "" {
				if healthCheckConf, err := util.PruneByTag(base.HealthCheckConf, "true", base.HealthCheckType); err == nil {
					base.HealthCheckConf = healthCheckConf.(HealthCheckConf)
				}
			}
			// Proxy type
			if p, err := util.PruneByTag(*proxy, "true", proxy.Type); err == nil {
				*proxy = p.(Proxy)
			}
			proxy.BaseProxyConf = base
		}
	}
}

func UnmarshalClientConfFromIni(source string) (*ClientConfig, error) {
	conf := NewDefaultClientConfig()
	cfg, err := ini.Load(source)
	if err != nil {
		return nil, err
	}
	// Load common options
	common, err := cfg.GetSection("common")
	if err != nil {
		return nil, err
	}
	if err = common.MapTo(&conf.ClientCommon); err != nil {
		return nil, err
	}
	// Load unparsed options
	conf.ClientCommon.Custom = make(map[string]string)
	for _, key := range common.Keys() {
		if util.GetFieldNameByTag(ClientCommon{}, "ini", key.Name()) == "" {
			conf.ClientCommon.Custom[key.Name()] = key.String()
		}
	}
	// Load all proxies
	for _, section := range cfg.Sections() {
		name := section.Name()
		if name == ini.DefaultSection || name == "common" {
			continue
		}
		proxy := Proxy{BaseProxyConf: BaseProxyConf{Name: name}}
		if err = section.MapTo(&proxy); err != nil {
			return nil, err
		}
		proxy.Custom = make(map[string]string)
		for _, key := range section.Keys() {
			if util.GetFieldNameByTag(Proxy{}, "ini", key.Name()) == "" {
				proxy.Custom[key.Name()] = key.String()
			}
		}
		conf.Proxies = append(conf.Proxies, &proxy)
	}
	conf.Complete()
	return conf, nil
}

func NewDefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		ClientCommon: ClientCommon{
			ClientAuth: ClientAuth{AuthMethod: "token"},
			ServerPort: "7000",
			LogLevel:   "info",
			LogMaxDays: 3,
		},
		Proxies: make([]*Proxy, 0),
	}
}
