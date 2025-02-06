package config

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/config/v1"
	frputil "github.com/fatedier/frp/pkg/util/util"
	"github.com/pelletier/go-toml/v2"
	"github.com/samber/lo"
	"gopkg.in/ini.v1"

	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/util"
)

type ClientAuth struct {
	AuthMethod                   string            `ini:"authentication_method,omitempty"`
	AuthenticateHeartBeats       bool              `ini:"authenticate_heartbeats,omitempty" token:"true" oidc:"true"`
	AuthenticateNewWorkConns     bool              `ini:"authenticate_new_work_conns,omitempty" token:"true" oidc:"true"`
	Token                        string            `ini:"token,omitempty" token:"true"`
	OIDCClientId                 string            `ini:"oidc_client_id,omitempty" oidc:"true"`
	OIDCClientSecret             string            `ini:"oidc_client_secret,omitempty" oidc:"true"`
	OIDCAudience                 string            `ini:"oidc_audience,omitempty" oidc:"true"`
	OIDCScope                    string            `ini:"oidc_scope,omitempty" oidc:"true"`
	OIDCTokenEndpoint            string            `ini:"oidc_token_endpoint_url,omitempty" oidc:"true"`
	OIDCAdditionalEndpointParams map[string]string `ini:"-" oidc:"true"`
}

func (ca ClientAuth) Complete() ClientAuth {
	authMethod := ca.AuthMethod
	if authMethod != "" {
		if auth, err := util.PruneByTag(ca, "true", authMethod); err == nil {
			ca = auth.(ClientAuth)
			ca.AuthMethod = authMethod
		}
		// Check the default auth method
		if authMethod == consts.AuthToken && ca.Token == "" {
			ca.AuthMethod = ""
		}
	} else {
		ca = ClientAuth{}
	}
	return ca
}

type ClientCommon struct {
	v1.APIMetadata            `ini:"-"`
	ClientAuth                `ini:",extends"`
	ServerAddress             string       `ini:"server_addr,omitempty"`
	ServerPort                int          `ini:"server_port,omitempty"`
	NatHoleSTUNServer         string       `ini:"nat_hole_stun_server,omitempty"`
	DialServerTimeout         int64        `ini:"dial_server_timeout,omitempty"`
	DialServerKeepAlive       int64        `ini:"dial_server_keepalive,omitempty"`
	ConnectServerLocalIP      string       `ini:"connect_server_local_ip,omitempty"`
	HTTPProxy                 string       `ini:"http_proxy,omitempty"`
	LogFile                   string       `ini:"log_file,omitempty"`
	LogLevel                  string       `ini:"log_level,omitempty"`
	LogMaxDays                int64        `ini:"log_max_days,omitempty"`
	AdminAddr                 string       `ini:"admin_addr,omitempty"`
	AdminPort                 int          `ini:"admin_port,omitempty"`
	AdminUser                 string       `ini:"admin_user,omitempty"`
	AdminPwd                  string       `ini:"admin_pwd,omitempty"`
	AdminTLS                  v1.TLSConfig `ini:"-"`
	AssetsDir                 string       `ini:"assets_dir,omitempty"`
	PoolCount                 int          `ini:"pool_count,omitempty"`
	DNSServer                 string       `ini:"dns_server,omitempty"`
	Protocol                  string       `ini:"protocol,omitempty"`
	QUICKeepalivePeriod       int          `ini:"quic_keepalive_period,omitempty"`
	QUICMaxIdleTimeout        int          `ini:"quic_max_idle_timeout,omitempty"`
	QUICMaxIncomingStreams    int          `ini:"quic_max_incoming_streams,omitempty"`
	LoginFailExit             bool         `ini:"login_fail_exit"`
	User                      string       `ini:"user,omitempty"`
	HeartbeatInterval         int64        `ini:"heartbeat_interval,omitempty"`
	HeartbeatTimeout          int64        `ini:"heartbeat_timeout,omitempty"`
	TCPMux                    bool         `ini:"tcp_mux"`
	TCPMuxKeepaliveInterval   int64        `ini:"tcp_mux_keepalive_interval,omitempty"`
	TLSEnable                 bool         `ini:"tls_enable"`
	TLSCertFile               string       `ini:"tls_cert_file,omitempty"`
	TLSKeyFile                string       `ini:"tls_key_file,omitempty"`
	TLSTrustedCaFile          string       `ini:"tls_trusted_ca_file,omitempty"`
	TLSServerName             string       `ini:"tls_server_name,omitempty"`
	UDPPacketSize             int64        `ini:"udp_packet_size,omitempty"`
	Start                     []string     `ini:"start,omitempty"`
	PprofEnable               bool         `ini:"pprof_enable,omitempty"`
	DisableCustomTLSFirstByte bool         `ini:"disable_custom_tls_first_byte"`

	// ManualStart defines whether to start the config on system boot.
	ManualStart bool `ini:"frpmgr_manual_start,omitempty"`
	// SVCBEnable resolves the SVCB record of server address.
	// By enabling this feature, the server address and port
	// can be updated dynamically.
	//
	// WARNING: This is an experimental feature.
	// It may affect the stability of the service.
	SVCBEnable bool `ini:"frpmgr_svcb_enable,omitempty"`
	// AutoDelete is a mechanism for temporary use.
	// The config will be stopped and deleted at some point.
	AutoDelete `ini:",extends"`
	// Client meta info
	Metas map[string]string `ini:"-"`
	// Config file format
	LegacyFormat bool `ini:"-"`
}

// BaseProxyConf provides configuration info that is common to all types.
type BaseProxyConf struct {
	// Name is the name of this proxy.
	Name string `ini:"-"`
	// Type specifies the type of this. Valid values include tcp, udp,
	// xtcp, stcp, sudp, http, https, tcpmux. By default, this value is "tcp".
	Type string `ini:"type,omitempty"`

	// UseEncryption controls whether communication with the server will
	// be encrypted. Encryption is done using the tokens supplied in the server
	// and client configuration. By default, this value is false.
	UseEncryption bool `ini:"use_encryption,omitempty"`
	// UseCompression controls whether communication with the server
	// will be compressed. By default, this value is false.
	UseCompression bool `ini:"use_compression,omitempty"`
	// Group specifies which group the proxy is a part of. The server will use
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

	// BandwidthLimit limits the bandwidth.
	// 0 means no limit.
	BandwidthLimit     string `ini:"bandwidth_limit,omitempty"`
	BandwidthLimitMode string `ini:"bandwidth_limit_mode,omitempty"`

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
	// Health checking parameters.
	HealthCheckConf `ini:",extends"`
	// Meta info for each proxy
	Metas map[string]string `ini:"-"`
	// Annotations for each proxy
	Annotations map[string]string `ini:"-"`
	// Disabled defines whether to start the proxy.
	Disabled bool `ini:"-"`
}

type PluginParams struct {
	PluginLocalAddr         string            `ini:"plugin_local_addr,omitempty" http2https:"true" http2http:"true" https2https:"true" https2http:"true" tls2raw:"true"`
	PluginCrtPath           string            `ini:"plugin_crt_path,omitempty" https2https:"true" https2http:"true" tls2raw:"true"`
	PluginKeyPath           string            `ini:"plugin_key_path,omitempty" https2https:"true" https2http:"true" tls2raw:"true"`
	PluginHostHeaderRewrite string            `ini:"plugin_host_header_rewrite,omitempty" http2https:"true" http2http:"true" https2https:"true" https2http:"true"`
	PluginHttpUser          string            `ini:"plugin_http_user,omitempty" http_proxy:"true" static_file:"true"`
	PluginHttpPasswd        string            `ini:"plugin_http_passwd,omitempty" http_proxy:"true" static_file:"true"`
	PluginUser              string            `ini:"plugin_user,omitempty" socks5:"true"`
	PluginPasswd            string            `ini:"plugin_passwd,omitempty" socks5:"true"`
	PluginLocalPath         string            `ini:"plugin_local_path,omitempty" static_file:"true"`
	PluginStripPrefix       string            `ini:"plugin_strip_prefix,omitempty" static_file:"true"`
	PluginUnixPath          string            `ini:"plugin_unix_path,omitempty" unix_domain_socket:"true"`
	PluginHeaders           map[string]string `ini:"-" http2https:"true" http2http:"true" https2https:"true" https2http:"true"`
	PluginEnableHTTP2       bool              `ini:"-" https2https:"true" https2http:"true"`
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
	// HealthCheckHTTPHeaders specifies the headers to send with the http request.
	HealthCheckHTTPHeaders map[string]string `ini:"-" http:"true"`
}

type Proxy struct {
	BaseProxyConf     `ini:",extends"`
	RemotePort        string            `ini:"remote_port,omitempty" tcp:"true" udp:"true"`
	Role              string            `ini:"role,omitempty" stcp:"true" xtcp:"true" sudp:"true" visitor:"*"`
	SK                string            `ini:"sk,omitempty" stcp:"true" xtcp:"true" sudp:"true" visitor:"*"`
	AllowUsers        string            `ini:"allow_users,omitempty" stcp:"true" xtcp:"true" sudp:"true"`
	ServerUser        string            `ini:"server_user,omitempty" visitor:"*"`
	ServerName        string            `ini:"server_name,omitempty" visitor:"*"`
	BindAddr          string            `ini:"bind_addr,omitempty" visitor:"*"`
	BindPort          int               `ini:"bind_port,omitempty" visitor:"*"`
	CustomDomains     string            `ini:"custom_domains,omitempty" http:"true" https:"true" tcpmux:"true"`
	SubDomain         string            `ini:"subdomain,omitempty" http:"true" https:"true" tcpmux:"true"`
	Locations         string            `ini:"locations,omitempty" http:"true"`
	HTTPUser          string            `ini:"http_user,omitempty" http:"true" tcpmux:"true"`
	HTTPPwd           string            `ini:"http_pwd,omitempty" http:"true" tcpmux:"true"`
	HostHeaderRewrite string            `ini:"host_header_rewrite,omitempty" http:"true"`
	Headers           map[string]string `ini:"-" http:"true"`
	ResponseHeaders   map[string]string `ini:"-" http:"true"`
	Multiplexer       string            `ini:"multiplexer,omitempty" tcpmux:"true"`
	RouteByHTTPUser   string            `ini:"route_by_http_user,omitempty" http:"true" tcpmux:"true"`
	// "kcp" or "quic"
	Protocol          string `ini:"protocol,omitempty" visitor:"xtcp"`
	KeepTunnelOpen    bool   `ini:"keep_tunnel_open,omitempty" visitor:"xtcp"`
	MaxRetriesAnHour  int    `ini:"max_retries_an_hour,omitempty" visitor:"xtcp"`
	MinRetryInterval  int    `ini:"min_retry_interval,omitempty" visitor:"xtcp"`
	FallbackTo        string `ini:"fallback_to,omitempty" visitor:"xtcp"`
	FallbackTimeoutMs int    `ini:"fallback_timeout_ms,omitempty" visitor:"xtcp"`
}

// GetAlias returns the alias of this proxy.
// It's usually equal to the proxy name, but proxies that start with "range:" differ from it.
func (p *Proxy) GetAlias() []string {
	if p.IsRange() {
		localPorts, err := frputil.ParseRangeNumbers(p.LocalPort)
		if err != nil {
			return []string{p.Name}
		}
		alias := make([]string, len(localPorts))
		for i := range localPorts {
			alias[i] = fmt.Sprintf("%s_%d", p.Name, i)
		}
		return alias
	}
	return []string{p.Name}
}

// IsVisitor returns a boolean indicating whether the proxy has a visitor role.
func (p *Proxy) IsVisitor() bool {
	return (p.Type == consts.ProxyTypeXTCP ||
		p.Type == consts.ProxyTypeSTCP ||
		p.Type == consts.ProxyTypeSUDP) && p.Role == "visitor"
}

func (p *Proxy) IsRange() bool {
	return (p.Type == consts.ProxyTypeTCP || p.Type == consts.ProxyTypeUDP) &&
		lo.Some([]rune(p.LocalPort+p.RemotePort), []rune{',', '-'})
}

// Complete removes redundant parameters base on the proxy type.
func (p *Proxy) Complete() {
	var base = p.BaseProxyConf
	if p.IsVisitor() {
		// Visitor
		if vp, err := util.PruneByTag(*p, p.Type, "visitor"); err == nil {
			*p = vp.(Proxy)
		}
		p.BaseProxyConf = BaseProxyConf{
			Name: base.Name, Type: base.Type, UseEncryption: base.UseEncryption,
			UseCompression: base.UseCompression, Disabled: base.Disabled,
		}
		// Reset xtcp visitor parameters
		if !p.KeepTunnelOpen {
			p.MaxRetriesAnHour = 0
			p.MinRetryInterval = 0
		}
		if p.FallbackTo == "" {
			p.FallbackTimeoutMs = 0
		}
	} else {
		// Plugins
		if base.Plugin != "" {
			base.LocalIP = ""
			base.LocalPort = ""
			if pluginParams, err := util.PruneByTag(base.PluginParams, "true", base.Plugin); err == nil {
				base.PluginParams = pluginParams.(PluginParams)
			}
		} else {
			base.PluginParams = PluginParams{}
		}
		// Health Check
		if base.HealthCheckType != "" {
			if healthCheckConf, err := util.PruneByTag(base.HealthCheckConf, "true", base.HealthCheckType); err == nil {
				base.HealthCheckConf = healthCheckConf.(HealthCheckConf)
			}
		} else {
			base.HealthCheckConf = HealthCheckConf{}
		}
		// Proxy type
		if typedProxy, err := util.PruneByTag(*p, "true", p.Type); err == nil {
			*p = typedProxy.(Proxy)
		}
		p.BaseProxyConf = base
	}
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

func (conf *ClientConfig) GetSTUNServer() string {
	return conf.NatHoleSTUNServer
}

func (conf *ClientConfig) Expiry() bool {
	switch conf.DeleteMethod {
	case consts.DeleteAbsolute:
		return true
	case consts.DeleteRelative:
		return conf.DeleteAfterDays > 0
	}
	return false
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
		if !slices.ContainsFunc(conf.Proxies, func(p *Proxy) bool { return p.Name == proxy.Name }) {
			conf.Proxies = append(conf.Proxies, proxy)
			return true
		}
	}
	return false
}

func (conf *ClientConfig) Save(path string) error {
	if conf.LegacyFormat {
		return conf.saveINI(path)
	} else {
		return conf.saveTOML(path)
	}
}

func (conf *ClientConfig) saveINI(path string) error {
	cfg := ini.Empty()
	common, err := cfg.NewSection("common")
	if err != nil {
		return err
	}
	if err = common.ReflectFrom(&conf.ClientCommon); err != nil {
		return err
	}
	for k, v := range conf.ClientCommon.Metas {
		common.Key("meta_" + k).SetValue(v)
	}
	for k, v := range conf.OIDCAdditionalEndpointParams {
		common.Key("oidc_additional_" + k).SetValue(v)
	}
	for _, proxy := range conf.Proxies {
		name := proxy.Name
		if proxy.IsRange() && !strings.HasPrefix(name, consts.RangePrefix) {
			name = consts.RangePrefix + name
		}
		p, err := cfg.NewSection(name)
		if err != nil {
			return err
		}
		if err = p.ReflectFrom(&proxy); err != nil {
			return err
		}
		for k, v := range proxy.Metas {
			p.Key("meta_" + k).SetValue(v)
		}
		for k, v := range proxy.Headers {
			p.Key("header_" + k).SetValue(v)
		}
		for k, v := range proxy.PluginHeaders {
			p.Key("plugin_header_" + k).SetValue(v)
		}
	}
	return cfg.SaveTo(path)
}

func (conf *ClientConfig) saveTOML(path string) error {
	c := ClientConfigV1{
		ClientCommonConfig: ClientCommonToV1(&conf.ClientCommon),
		Mgr: Mgr{
			ManualStart: conf.ManualStart,
			SVCBEnable:  conf.SVCBEnable,
			AutoDelete:  conf.AutoDelete,
		},
	}
	for i, v := range conf.Proxies {
		if v.IsVisitor() {
			visitor := ClientVisitorToV1(v)
			visitor.Mgr.Sort = i + 1
			c.Visitors = append(c.Visitors, visitor)
		} else {
			proxies, err := ClientProxyToV1(v)
			if err != nil {
				return err
			}
			c.Proxies = append(c.Proxies, proxies...)
		}
	}
	obj, err := toMap(&c, "json")
	if err != nil {
		return err
	}
	b, err := toml.Marshal(obj)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0666)
}

func (conf *ClientConfig) Complete(read bool) {
	// Common config
	conf.ClientAuth = conf.ClientAuth.Complete()
	if conf.AdminPort == 0 {
		conf.AdminUser = ""
		conf.AdminPwd = ""
		conf.AssetsDir = ""
		conf.AdminTLS = v1.TLSConfig{}
		conf.PprofEnable = false
	}
	conf.AutoDelete = conf.AutoDelete.Complete()
	if !conf.TCPMux {
		conf.TCPMuxKeepaliveInterval = 0
	}
	if !conf.TLSEnable {
		conf.TLSServerName = ""
		conf.TLSCertFile = ""
		conf.TLSKeyFile = ""
		conf.TLSTrustedCaFile = ""
	}
	if conf.Protocol == consts.ProtoQUIC {
		conf.DialServerTimeout = 0
		conf.DialServerKeepAlive = 0
	} else {
		conf.QUICMaxIdleTimeout = 0
		conf.QUICKeepalivePeriod = 0
		conf.QUICMaxIncomingStreams = 0
	}
	// Proxies
	for _, proxy := range conf.Proxies {
		// Complete proxy
		proxy.Complete()
		// Check proxy status
		if read && len(conf.Start) > 0 {
			proxy.Disabled = !lo.Every(conf.Start, proxy.GetAlias())
		}
	}
	if !read {
		conf.Start = conf.gatherStart()
	}
}

func (conf *ClientConfig) Copy(all bool) Config {
	newConf := NewDefaultClientConfig()
	newConf.ClientCommon = conf.ClientCommon
	// We can't share the same log file between different configs
	newConf.ClientCommon.LogFile = ""
	if all {
		for _, proxy := range conf.Proxies {
			var newProxy = *proxy
			newConf.Proxies = append(newConf.Proxies, &newProxy)
		}
	}
	return newConf
}

// gatherStart returns a list of enabled proxies name, or a nil slice if all proxies are enabled.
func (conf *ClientConfig) gatherStart() []string {
	allStart := true
	start := make([]string, 0)
	for _, proxy := range conf.Proxies {
		if !proxy.Disabled {
			start = append(start, proxy.GetAlias()...)
		} else {
			allStart = false
		}
	}
	if allStart {
		return nil
	}
	return start
}

// CountStart returns the number of enabled proxies.
func (conf *ClientConfig) CountStart() int {
	return len(lo.Filter(conf.Proxies, func(proxy *Proxy, i int) bool { return !proxy.Disabled }))
}

func (conf *ClientConfig) Ext() string {
	if conf.LegacyFormat {
		return ".ini"
	} else {
		return ".toml"
	}
}

// NewProxyFromIni creates a proxy object from ini section
func NewProxyFromIni(name string, section *ini.Section) (*Proxy, error) {
	proxy := NewDefaultProxyConfig(name)
	if err := section.MapTo(&proxy); err != nil {
		return nil, err
	}
	proxy.Metas = util.GetMapWithoutPrefix(section.KeysHash(), "meta_")
	proxy.Headers = util.GetMapWithoutPrefix(section.KeysHash(), "header_")
	proxy.PluginHeaders = util.GetMapWithoutPrefix(section.KeysHash(), "plugin_header_")
	proxy.Name = strings.TrimPrefix(proxy.Name, consts.RangePrefix)
	return proxy, nil
}

// UnmarshalProxyFromIni finds a single proxy section and unmarshals it from ini source.
func UnmarshalProxyFromIni(source interface{}) (*Proxy, error) {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
		AllowBooleanKeys:    true,
	}, source)
	if err != nil {
		return nil, err
	}
	var useName string
	var useSection *ini.Section
	// Try to find a proxy section
findSection:
	for _, section := range cfg.Sections() {
		switch section.Name() {
		case "common":
			continue
		case ini.DefaultSection:
			// Use the default section if no proxy is found
			useName, useSection = "", section
			continue
		default:
			useName, useSection = section.Name(), section
			break findSection
		}
	}
	if useSection == nil || len(useSection.Keys()) == 0 {
		return nil, ini.ErrDelimiterNotFound{}
	}
	return NewProxyFromIni(useName, useSection)
}

func UnmarshalClientConfFromIni(source interface{}) (*ClientConfig, error) {
	conf := NewDefaultClientConfig()
	cfg, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
		AllowBooleanKeys:    true,
	}, source)
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
	conf.Metas = util.GetMapWithoutPrefix(common.KeysHash(), "meta_")
	conf.OIDCAdditionalEndpointParams = util.GetMapWithoutPrefix(common.KeysHash(), "oidc_additional_")
	// Load all proxies
	for _, section := range cfg.Sections() {
		name := section.Name()
		if name == ini.DefaultSection || name == "common" {
			continue
		}
		proxy, err := NewProxyFromIni(name, section)
		if err != nil {
			return nil, err
		}
		conf.Proxies = append(conf.Proxies, proxy)
	}
	conf.Complete(true)
	conf.LegacyFormat = true
	return conf, nil
}

func UnmarshalClientConf(source interface{}) (*ClientConfig, error) {
	var b []byte
	var err error
	if path, ok := source.(string); ok {
		b, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	} else {
		b = source.([]byte)
	}
	if config.DetectLegacyINIFormat(b) {
		return UnmarshalClientConfFromIni(source)
	}
	var cfg = NewDefaultClientConfigV1()
	if err = config.LoadConfigure(b, &cfg, false); err != nil {
		return nil, err
	}
	var conf ClientConfig
	conf.ClientCommon = ClientCommonFromV1(&cfg.ClientCommonConfig)
	conf.ManualStart = cfg.Mgr.ManualStart
	conf.SVCBEnable = cfg.Mgr.SVCBEnable
	conf.AutoDelete = cfg.Mgr.AutoDelete
	// Proxies
	ignore := make(map[string]struct{})
	proxies := make([]*Proxy, len(cfg.Proxies))
	for i, v := range cfg.Proxies {
		p := ClientProxyFromV1(v)
		if p.IsRange() {
			for _, name := range p.GetAlias() {
				if name != p.Name {
					ignore[name] = struct{}{}
				}
			}
		}
		proxies[i] = p
	}
	conf.Proxies = lo.Filter(proxies, func(item *Proxy, index int) bool {
		_, ok := ignore[item.Name]
		return !ok
	})
	// Visitors
	slices.SortStableFunc(cfg.Visitors, func(a, b TypedVisitorConfig) int {
		if a.Mgr.Sort <= 0 && b.Mgr.Sort <= 0 {
			return 0
		}
		return a.Mgr.Sort - b.Mgr.Sort
	})
	for _, v := range cfg.Visitors {
		visitor := ClientVisitorFromV1(v)
		if v.Mgr.Sort <= 0 {
			conf.Proxies = append(conf.Proxies, visitor)
		} else {
			conf.Proxies = slices.Insert(conf.Proxies, min(v.Mgr.Sort-1, len(conf.Proxies)), visitor)
		}
	}
	conf.Complete(true)
	return &conf, nil
}

func NewDefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		ClientCommon: ClientCommon{
			ClientAuth:                ClientAuth{AuthMethod: consts.AuthToken},
			ServerPort:                consts.DefaultServerPort,
			LogLevel:                  consts.LogLevelInfo,
			LogMaxDays:                consts.DefaultLogMaxDays,
			TCPMux:                    true,
			TLSEnable:                 true,
			DisableCustomTLSFirstByte: true,
			AutoDelete:                AutoDelete{DeleteMethod: consts.DeleteRelative},
		},
		Proxies: make([]*Proxy, 0),
	}
}

func NewDefaultClientConfigV1() ClientConfigV1 {
	return ClientConfigV1{
		ClientCommonConfig: v1.ClientCommonConfig{
			Auth:          v1.AuthClientConfig{Method: v1.AuthMethodToken},
			ServerPort:    consts.DefaultServerPort,
			Log:           v1.LogConfig{Level: consts.LogLevelInfo, MaxDays: consts.DefaultLogMaxDays},
			LoginFailExit: lo.ToPtr(false),
		},
		Mgr: Mgr{AutoDelete: AutoDelete{DeleteMethod: consts.DeleteRelative}},
	}
}

func NewDefaultProxyConfig(name string) *Proxy {
	return &Proxy{
		BaseProxyConf: BaseProxyConf{
			Name: name, Type: consts.ProxyTypeTCP,
		},
	}
}
