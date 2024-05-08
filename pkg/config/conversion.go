package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/fatedier/frp/pkg/config/types"
	"github.com/fatedier/frp/pkg/config/v1"
	frputil "github.com/fatedier/frp/pkg/util/util"
	"github.com/samber/lo"

	"github.com/koho/frpmgr/pkg/consts"
)

func ClientCommonFromV1(c *v1.ClientCommonConfig) (r ClientCommon) {
	r.APIMetadata = c.APIMetadata

	// Auth client config
	r.AuthMethod = string(c.Auth.Method)
	r.Token = c.Auth.Token
	r.OIDCClientId = c.Auth.OIDC.ClientID
	r.OIDCClientSecret = c.Auth.OIDC.ClientSecret
	r.OIDCAudience = c.Auth.OIDC.Audience
	r.OIDCScope = c.Auth.OIDC.Scope
	r.OIDCTokenEndpoint = c.Auth.OIDC.TokenEndpointURL
	r.OIDCAdditionalEndpointParams = c.Auth.OIDC.AdditionalEndpointParams
	if lo.Contains(c.Auth.AdditionalScopes, v1.AuthScopeHeartBeats) {
		r.AuthenticateHeartBeats = true
	}
	if lo.Contains(c.Auth.AdditionalScopes, v1.AuthScopeNewWorkConns) {
		r.AuthenticateNewWorkConns = true
	}

	r.User = c.User
	r.ServerAddress = c.ServerAddr
	r.ServerPort = c.ServerPort
	r.NatHoleSTUNServer = c.NatHoleSTUNServer
	r.DNSServer = c.DNSServer
	if c.LoginFailExit == nil || *c.LoginFailExit {
		r.LoginFailExit = true
	}
	r.Start = c.Start

	// Log
	r.LogFile = c.Log.To
	r.LogLevel = c.Log.Level
	r.LogMaxDays = c.Log.MaxDays

	// Admin
	r.AdminAddr = c.WebServer.Addr
	r.AdminPort = c.WebServer.Port
	r.AdminUser = c.WebServer.User
	r.AdminPwd = c.WebServer.Password
	r.AssetsDir = c.WebServer.AssetsDir
	r.PprofEnable = c.WebServer.PprofEnable
	if c.WebServer.TLS != nil {
		r.AdminTLS = *c.WebServer.TLS
	}

	// Transport
	r.Protocol = c.Transport.Protocol
	r.DialServerTimeout = c.Transport.DialServerTimeout
	r.DialServerKeepAlive = c.Transport.DialServerKeepAlive
	r.ConnectServerLocalIP = c.Transport.ConnectServerLocalIP
	r.HTTPProxy = c.Transport.ProxyURL
	r.PoolCount = c.Transport.PoolCount
	if c.Transport.TCPMux == nil || *c.Transport.TCPMux {
		r.TCPMux = true
	}
	r.TCPMuxKeepaliveInterval = c.Transport.TCPMuxKeepaliveInterval
	r.QUICMaxIncomingStreams = c.Transport.QUIC.MaxIncomingStreams
	r.QUICKeepalivePeriod = c.Transport.QUIC.KeepalivePeriod
	r.QUICMaxIdleTimeout = c.Transport.QUIC.MaxIdleTimeout
	r.HeartbeatInterval = c.Transport.HeartbeatInterval
	r.HeartbeatTimeout = c.Transport.HeartbeatTimeout
	if c.Transport.TLS.Enable == nil || *c.Transport.TLS.Enable {
		r.TLSEnable = true
	}
	r.TLSCertFile = c.Transport.TLS.CertFile
	r.TLSKeyFile = c.Transport.TLS.KeyFile
	r.TLSTrustedCaFile = c.Transport.TLS.TrustedCaFile
	r.TLSServerName = c.Transport.TLS.ServerName
	if c.Transport.TLS.DisableCustomTLSFirstByte == nil || *c.Transport.TLS.DisableCustomTLSFirstByte {
		r.DisableCustomTLSFirstByte = true
	}
	r.UDPPacketSize = c.UDPPacketSize
	r.Metas = c.Metadatas
	return
}

func ClientProxyFromV1(pxyCfg TypedProxyConfig) *Proxy {
	var r Proxy
	clientProxyBaseFromV1(pxyCfg.GetBaseConfig(), &r)
	setRemotePort := func(port int) {
		if pxyCfg.Mgr.Range.Local != "" && pxyCfg.Mgr.Range.Remote != "" && strings.HasSuffix(r.Name, "_0") {
			r.Name = strings.TrimSuffix(r.Name, "_0")
			r.LocalPort = pxyCfg.Mgr.Range.Local
			r.RemotePort = pxyCfg.Mgr.Range.Remote
		} else {
			r.RemotePort = strconv.Itoa(port)
		}
	}
	switch v := pxyCfg.ProxyConfigurer.(type) {
	case *v1.TCPProxyConfig:
		setRemotePort(v.RemotePort)
	case *v1.UDPProxyConfig:
		setRemotePort(v.RemotePort)
	case *v1.HTTPProxyConfig:
		r.SubDomain = v.SubDomain
		r.CustomDomains = strings.Join(v.CustomDomains, ",")
		r.Locations = strings.Join(v.Locations, ",")
		r.HTTPUser = v.HTTPUser
		r.HTTPPwd = v.HTTPPassword
		r.HostHeaderRewrite = v.HostHeaderRewrite
		r.RouteByHTTPUser = v.RouteByHTTPUser
		r.Headers = v.RequestHeaders.Set
		r.ResponseHeaders = v.ResponseHeaders.Set
	case *v1.HTTPSProxyConfig:
		r.SubDomain = v.SubDomain
		r.CustomDomains = strings.Join(v.CustomDomains, ",")
	case *v1.TCPMuxProxyConfig:
		r.SubDomain = v.SubDomain
		r.CustomDomains = strings.Join(v.CustomDomains, ",")
		r.HTTPUser = v.HTTPUser
		r.HTTPPwd = v.HTTPPassword
		r.RouteByHTTPUser = v.RouteByHTTPUser
		r.Multiplexer = v.Multiplexer
	case *v1.STCPProxyConfig:
		r.SK = v.Secretkey
		r.AllowUsers = strings.Join(v.AllowUsers, ",")
	case *v1.SUDPProxyConfig:
		r.SK = v.Secretkey
		r.AllowUsers = strings.Join(v.AllowUsers, ",")
	case *v1.XTCPProxyConfig:
		r.SK = v.Secretkey
		r.AllowUsers = strings.Join(v.AllowUsers, ",")
	}
	return &r
}

func ClientVisitorFromV1(visitorCfg v1.TypedVisitorConfig) *Proxy {
	var r Proxy
	clientVisitorBaseFromV1(visitorCfg.GetBaseConfig(), &r)
	switch v := visitorCfg.VisitorConfigurer.(type) {
	case *v1.STCPVisitorConfig:
	case *v1.SUDPVisitorConfig:
	case *v1.XTCPVisitorConfig:
		r.Protocol = v.Protocol
		r.KeepTunnelOpen = v.KeepTunnelOpen
		r.MaxRetriesAnHour = v.MaxRetriesAnHour
		r.MinRetryInterval = v.MinRetryInterval
		r.FallbackTo = v.FallbackTo
		r.FallbackTimeoutMs = v.FallbackTimeoutMs
	}
	return &r
}

func clientProxyBaseFromV1(c *v1.ProxyBaseConfig, out *Proxy) {
	out.Name = c.Name
	out.Type = c.Type
	out.UseEncryption = c.Transport.UseEncryption
	out.UseCompression = c.Transport.UseCompression
	out.BandwidthLimitMode = c.Transport.BandwidthLimitMode
	out.BandwidthLimit = c.Transport.BandwidthLimit.String()
	out.ProxyProtocolVersion = c.Transport.ProxyProtocolVersion
	out.Group = c.LoadBalancer.Group
	out.GroupKey = c.LoadBalancer.GroupKey
	out.HealthCheckType = c.HealthCheck.Type
	out.HealthCheckTimeoutS = c.HealthCheck.TimeoutSeconds
	out.HealthCheckMaxFailed = c.HealthCheck.MaxFailed
	out.HealthCheckIntervalS = c.HealthCheck.IntervalSeconds
	out.HealthCheckURL = c.HealthCheck.Path
	out.HealthCheckHTTPHeaders = lo.SliceToMap(c.HealthCheck.HTTPHeaders, func(item v1.HTTPHeader) (string, string) {
		return item.Name, item.Value
	})
	out.LocalIP = c.LocalIP
	if c.LocalPort != 0 {
		out.LocalPort = strconv.Itoa(c.LocalPort)
	}

	out.Metas = c.Metadatas
	out.Annotations = c.Annotations
	out.Plugin = c.Plugin.Type
	switch v := c.Plugin.ClientPluginOptions.(type) {
	case *v1.HTTP2HTTPSPluginOptions:
		out.PluginLocalAddr = v.LocalAddr
		out.PluginHostHeaderRewrite = v.HostHeaderRewrite
		out.PluginHeaders = v.RequestHeaders.Set
	case *v1.HTTPProxyPluginOptions:
		out.PluginHttpUser = v.HTTPUser
		out.PluginHttpPasswd = v.HTTPPassword
	case *v1.HTTPS2HTTPPluginOptions:
		out.PluginLocalAddr = v.LocalAddr
		out.PluginHostHeaderRewrite = v.HostHeaderRewrite
		out.PluginCrtPath = v.CrtPath
		out.PluginKeyPath = v.KeyPath
		out.PluginHeaders = v.RequestHeaders.Set
	case *v1.HTTPS2HTTPSPluginOptions:
		out.PluginLocalAddr = v.LocalAddr
		out.PluginHostHeaderRewrite = v.HostHeaderRewrite
		out.PluginCrtPath = v.CrtPath
		out.PluginKeyPath = v.KeyPath
		out.PluginHeaders = v.RequestHeaders.Set
	case *v1.Socks5PluginOptions:
		out.PluginUser = v.Username
		out.PluginPasswd = v.Password
	case *v1.StaticFilePluginOptions:
		out.PluginLocalPath = v.LocalPath
		out.PluginStripPrefix = v.StripPrefix
		out.PluginHttpUser = v.HTTPUser
		out.PluginHttpPasswd = v.HTTPPassword
	case *v1.UnixDomainSocketPluginOptions:
		out.PluginUnixPath = v.UnixPath
	}
}

func clientVisitorBaseFromV1(c *v1.VisitorBaseConfig, out *Proxy) {
	out.Name = c.Name
	out.Type = c.Type
	out.Role = "visitor"
	out.UseEncryption = c.Transport.UseEncryption
	out.UseCompression = c.Transport.UseCompression
	out.SK = c.SecretKey
	out.ServerUser = c.ServerUser
	out.ServerName = c.ServerName
	out.BindAddr = c.BindAddr
	out.BindPort = c.BindPort
}

func ClientCommonToV1(c *ClientCommon) (r v1.ClientCommonConfig) {
	r.APIMetadata = c.APIMetadata

	// Auth client config
	r.Auth = v1.AuthClientConfig{
		Method: v1.AuthMethod(c.AuthMethod),
		Token:  c.Token,
		OIDC: v1.AuthOIDCClientConfig{
			ClientID:                 c.OIDCClientId,
			ClientSecret:             c.OIDCClientSecret,
			Audience:                 c.OIDCAudience,
			Scope:                    c.OIDCScope,
			TokenEndpointURL:         c.OIDCTokenEndpoint,
			AdditionalEndpointParams: c.OIDCAdditionalEndpointParams,
		},
	}
	if c.AuthenticateHeartBeats {
		r.Auth.AdditionalScopes = append(r.Auth.AdditionalScopes, v1.AuthScopeHeartBeats)
	}
	if c.AuthenticateNewWorkConns {
		r.Auth.AdditionalScopes = append(r.Auth.AdditionalScopes, v1.AuthScopeNewWorkConns)
	}

	r.User = c.User
	r.ServerAddr = c.ServerAddress
	r.ServerPort = c.ServerPort
	r.NatHoleSTUNServer = c.NatHoleSTUNServer
	r.DNSServer = c.DNSServer
	r.LoginFailExit = &c.LoginFailExit
	r.Start = c.Start

	// Log
	r.Log = v1.LogConfig{
		To:      c.LogFile,
		Level:   c.LogLevel,
		MaxDays: c.LogMaxDays,
	}

	// Admin
	r.WebServer = v1.WebServerConfig{
		Addr:        c.AdminAddr,
		Port:        c.AdminPort,
		User:        c.AdminUser,
		Password:    c.AdminPwd,
		AssetsDir:   c.AssetsDir,
		PprofEnable: c.PprofEnable,
	}
	if lo.IsNotEmpty(c.AdminTLS) {
		r.WebServer.TLS = &c.AdminTLS
	}

	// Transport
	r.Transport = v1.ClientTransportConfig{
		Protocol:                c.Protocol,
		DialServerTimeout:       c.DialServerTimeout,
		DialServerKeepAlive:     c.DialServerKeepAlive,
		ConnectServerLocalIP:    c.ConnectServerLocalIP,
		ProxyURL:                c.HTTPProxy,
		PoolCount:               c.PoolCount,
		TCPMux:                  &c.TCPMux,
		TCPMuxKeepaliveInterval: c.TCPMuxKeepaliveInterval,
		QUIC: v1.QUICOptions{
			KeepalivePeriod:    c.QUICKeepalivePeriod,
			MaxIdleTimeout:     c.QUICMaxIdleTimeout,
			MaxIncomingStreams: c.QUICMaxIncomingStreams,
		},
		HeartbeatInterval: c.HeartbeatInterval,
		HeartbeatTimeout:  c.HeartbeatTimeout,
		TLS: v1.TLSClientConfig{
			Enable:                    &c.TLSEnable,
			DisableCustomTLSFirstByte: &c.DisableCustomTLSFirstByte,
			TLSConfig: v1.TLSConfig{
				CertFile:      c.TLSCertFile,
				KeyFile:       c.TLSKeyFile,
				TrustedCaFile: c.TLSTrustedCaFile,
				ServerName:    c.TLSServerName,
			},
		},
	}
	r.UDPPacketSize = c.UDPPacketSize
	r.Metadatas = c.Metas
	return
}

func ClientProxyToV1(p *Proxy) ([]TypedProxyConfig, error) {
	if p.IsRange() {
		localPorts, err := frputil.ParseRangeNumbers(p.LocalPort)
		if err != nil {
			return nil, err
		}
		remotePorts, err := frputil.ParseRangeNumbers(p.RemotePort)
		if err != nil {
			return nil, err
		}
		if len(localPorts) != len(remotePorts) {
			return nil, fmt.Errorf("local ports number should be same with remote ports number")
		}
		r := make([]TypedProxyConfig, len(localPorts))
		for i := range localPorts {
			subPxy := *p
			subPxy.Name = fmt.Sprintf("%s_%d", p.Name, i)
			subPxy.LocalPort = strconv.FormatInt(localPorts[i], 10)
			subPxy.RemotePort = strconv.FormatInt(remotePorts[i], 10)
			if r[i], err = singleClientProxyToV1(&subPxy); err != nil {
				return nil, err
			}
		}
		r[0].Mgr.Range.Local = p.LocalPort
		r[0].Mgr.Range.Remote = p.RemotePort
		return r, nil
	} else {
		r, err := singleClientProxyToV1(p)
		if err != nil {
			return nil, err
		}
		return []TypedProxyConfig{r}, nil
	}
}

func singleClientProxyToV1(p *Proxy) (TypedProxyConfig, error) {
	r := TypedProxyConfig{TypedProxyConfig: v1.TypedProxyConfig{Type: p.Type}}
	base, err := clientProxyBaseToV1(&p.BaseProxyConf)
	if err != nil {
		return r, err
	}
	switch r.Type {
	case consts.ProxyTypeTCP:
		c := &v1.TCPProxyConfig{ProxyBaseConfig: base}
		if p.RemotePort != "" {
			if c.RemotePort, err = strconv.Atoi(p.RemotePort); err != nil {
				return r, err
			}
		}
		r.ProxyConfigurer = c
	case consts.ProxyTypeUDP:
		c := &v1.UDPProxyConfig{ProxyBaseConfig: base}
		if p.RemotePort != "" {
			if c.RemotePort, err = strconv.Atoi(p.RemotePort); err != nil {
				return r, err
			}
		}
		r.ProxyConfigurer = c
	case consts.ProxyTypeHTTP:
		c := &v1.HTTPProxyConfig{
			ProxyBaseConfig: base,
			DomainConfig: v1.DomainConfig{
				SubDomain: p.SubDomain,
			},
			HTTPUser:          p.HTTPUser,
			HTTPPassword:      p.HTTPPwd,
			HostHeaderRewrite: p.HostHeaderRewrite,
			RequestHeaders: v1.HeaderOperations{
				Set: p.Headers,
			},
			ResponseHeaders: v1.HeaderOperations{
				Set: p.ResponseHeaders,
			},
			RouteByHTTPUser: p.RouteByHTTPUser,
		}
		if p.CustomDomains != "" {
			c.CustomDomains = strings.Split(p.CustomDomains, ",")
		}
		if p.Locations != "" {
			c.Locations = strings.Split(p.Locations, ",")
		}
		r.ProxyConfigurer = c
	case consts.ProxyTypeHTTPS:
		c := &v1.HTTPSProxyConfig{
			ProxyBaseConfig: base,
			DomainConfig: v1.DomainConfig{
				SubDomain: p.SubDomain,
			},
		}
		if p.CustomDomains != "" {
			c.CustomDomains = strings.Split(p.CustomDomains, ",")
		}
		r.ProxyConfigurer = c
	case consts.ProxyTypeTCPMUX:
		c := &v1.TCPMuxProxyConfig{
			ProxyBaseConfig: base,
			DomainConfig: v1.DomainConfig{
				SubDomain: p.SubDomain,
			},
			HTTPUser:        p.HTTPUser,
			HTTPPassword:    p.HTTPPwd,
			RouteByHTTPUser: p.RouteByHTTPUser,
			Multiplexer:     p.Multiplexer,
		}
		if p.CustomDomains != "" {
			c.CustomDomains = strings.Split(p.CustomDomains, ",")
		}
		r.ProxyConfigurer = c
	case consts.ProxyTypeSTCP:
		c := &v1.STCPProxyConfig{
			ProxyBaseConfig: base,
			Secretkey:       p.SK,
		}
		if p.AllowUsers != "" {
			c.AllowUsers = strings.Split(p.AllowUsers, ",")
		}
		r.ProxyConfigurer = c
	case consts.ProxyTypeSUDP:
		c := &v1.SUDPProxyConfig{
			ProxyBaseConfig: base,
			Secretkey:       p.SK,
		}
		if p.AllowUsers != "" {
			c.AllowUsers = strings.Split(p.AllowUsers, ",")
		}
		r.ProxyConfigurer = c
	case consts.ProxyTypeXTCP:
		c := &v1.XTCPProxyConfig{
			ProxyBaseConfig: base,
			Secretkey:       p.SK,
		}
		if p.AllowUsers != "" {
			c.AllowUsers = strings.Split(p.AllowUsers, ",")
		}
		r.ProxyConfigurer = c
	}
	return r, nil
}

func clientProxyBaseToV1(c *BaseProxyConf) (v1.ProxyBaseConfig, error) {
	r := v1.ProxyBaseConfig{
		Name: c.Name,
		Type: c.Type,
		Transport: v1.ProxyTransport{
			UseEncryption:        c.UseEncryption,
			UseCompression:       c.UseCompression,
			BandwidthLimitMode:   c.BandwidthLimitMode,
			ProxyProtocolVersion: c.ProxyProtocolVersion,
		},
		Metadatas:   c.Metas,
		Annotations: c.Annotations,
		LoadBalancer: v1.LoadBalancerConfig{
			Group:    c.Group,
			GroupKey: c.GroupKey,
		},
		HealthCheck: v1.HealthCheckConfig{
			Type:            c.HealthCheckType,
			TimeoutSeconds:  c.HealthCheckTimeoutS,
			MaxFailed:       c.HealthCheckMaxFailed,
			IntervalSeconds: c.HealthCheckIntervalS,
			Path:            c.HealthCheckURL,
			HTTPHeaders: lo.MapToSlice(c.HealthCheckHTTPHeaders, func(key string, value string) v1.HTTPHeader {
				return v1.HTTPHeader{Name: key, Value: value}
			}),
		},
		ProxyBackend: v1.ProxyBackend{
			LocalIP: c.LocalIP,
			Plugin: v1.TypedClientPluginOptions{
				Type: c.Plugin,
			},
		},
	}
	if c.LocalPort != "" {
		localPort, err := strconv.Atoi(c.LocalPort)
		if err != nil {
			return r, err
		}
		r.ProxyBackend.LocalPort = localPort
	}
	bl, err := types.NewBandwidthQuantity(c.BandwidthLimit)
	if err != nil {
		return r, err
	}
	r.Transport.BandwidthLimit = bl
	switch c.Plugin {
	case consts.PluginHttp2Https:
		r.ProxyBackend.Plugin.ClientPluginOptions = &v1.HTTP2HTTPSPluginOptions{
			Type:              c.Plugin,
			LocalAddr:         c.PluginLocalAddr,
			HostHeaderRewrite: c.PluginHostHeaderRewrite,
			RequestHeaders: v1.HeaderOperations{
				Set: c.PluginHeaders,
			},
		}
	case consts.PluginHttpProxy:
		r.ProxyBackend.Plugin.ClientPluginOptions = &v1.HTTPProxyPluginOptions{
			Type:         c.Plugin,
			HTTPUser:     c.PluginHttpUser,
			HTTPPassword: c.PluginHttpPasswd,
		}
	case consts.PluginHttps2Http:
		r.ProxyBackend.Plugin.ClientPluginOptions = &v1.HTTPS2HTTPPluginOptions{
			Type:              c.Plugin,
			LocalAddr:         c.PluginLocalAddr,
			HostHeaderRewrite: c.PluginHostHeaderRewrite,
			RequestHeaders: v1.HeaderOperations{
				Set: c.PluginHeaders,
			},
			CrtPath: c.PluginCrtPath,
			KeyPath: c.PluginKeyPath,
		}
	case consts.PluginHttps2Https:
		r.ProxyBackend.Plugin.ClientPluginOptions = &v1.HTTPS2HTTPSPluginOptions{
			Type:              c.Plugin,
			LocalAddr:         c.PluginLocalAddr,
			HostHeaderRewrite: c.PluginHostHeaderRewrite,
			RequestHeaders: v1.HeaderOperations{
				Set: c.PluginHeaders,
			},
			CrtPath: c.PluginCrtPath,
			KeyPath: c.PluginKeyPath,
		}
	case consts.PluginSocks5:
		r.ProxyBackend.Plugin.ClientPluginOptions = &v1.Socks5PluginOptions{
			Type:     c.Plugin,
			Username: c.PluginUser,
			Password: c.PluginPasswd,
		}
	case consts.PluginStaticFile:
		r.ProxyBackend.Plugin.ClientPluginOptions = &v1.StaticFilePluginOptions{
			Type:         c.Plugin,
			LocalPath:    c.PluginLocalPath,
			StripPrefix:  c.PluginStripPrefix,
			HTTPUser:     c.PluginHttpUser,
			HTTPPassword: c.PluginHttpPasswd,
		}
	case consts.PluginUnixDomain:
		r.ProxyBackend.Plugin.ClientPluginOptions = &v1.UnixDomainSocketPluginOptions{
			Type:     c.Plugin,
			UnixPath: c.PluginUnixPath,
		}
	}
	return r, nil
}

func ClientVisitorToV1(p *Proxy) v1.TypedVisitorConfig {
	r := v1.TypedVisitorConfig{Type: p.Type}
	base := clientVisitorBaseToV1(p)
	switch p.Type {
	case consts.ProxyTypeSTCP:
		r.VisitorConfigurer = &v1.STCPVisitorConfig{VisitorBaseConfig: base}
	case consts.ProxyTypeSUDP:
		r.VisitorConfigurer = &v1.SUDPVisitorConfig{VisitorBaseConfig: base}
	case consts.ProxyTypeXTCP:
		r.VisitorConfigurer = &v1.XTCPVisitorConfig{
			VisitorBaseConfig: base,
			Protocol:          p.Protocol,
			KeepTunnelOpen:    p.KeepTunnelOpen,
			MaxRetriesAnHour:  p.MaxRetriesAnHour,
			MinRetryInterval:  p.MinRetryInterval,
			FallbackTo:        p.FallbackTo,
			FallbackTimeoutMs: p.FallbackTimeoutMs,
		}
	}
	return r
}

func clientVisitorBaseToV1(p *Proxy) v1.VisitorBaseConfig {
	return v1.VisitorBaseConfig{
		Name: p.Name,
		Type: p.Type,
		Transport: v1.VisitorTransport{
			UseEncryption:  p.UseEncryption,
			UseCompression: p.UseCompression,
		},
		SecretKey:  p.SK,
		ServerUser: p.ServerUser,
		ServerName: p.ServerName,
		BindAddr:   p.BindAddr,
		BindPort:   p.BindPort,
	}
}

// toMap converts a struct to a map using the struct tags.
func toMap(in any, tag string) (map[string]any, error) {
	out := make(map[string]any)

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// we only accept structs
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("toMap only accepts structs; got %T", v)
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		ft := t.Field(i)
		fv := v.Field(i)
		key := ft.Tag.Get(tag)
		if key != "" {
			key = strings.Split(key, ",")[0]
		}
		if key == "-" {
			continue
		}
		switch fv.Kind() {
		case reflect.Struct, reflect.Interface:
			value := fv.Interface()
			if value == nil {
				continue
			}
			switch o := value.(type) {
			case time.Time:
				if !o.IsZero() && key != "" {
					if b, err := o.MarshalText(); err != nil {
						return nil, err
					} else {
						out[key] = string(b)
					}
				}
			case types.BandwidthQuantity:
				if o.String() != "" && key != "" {
					out[key] = o.String()
				}
			default:
				m, err := toMap(value, tag)
				if err != nil {
					return nil, err
				}
				if len(m) == 0 {
					continue
				}
				if key == "" {
					if ft.Anonymous {
						for k, v := range m {
							out[k] = v
						}
					}
				} else {
					out[key] = m
				}
			}
		case reflect.Slice:
			if key == "" || fv.Len() == 0 {
				continue
			}
			if fv.Index(0).Kind() == reflect.Struct {
				s := make([]any, fv.Len())
				for k := 0; k < fv.Len(); k++ {
					m, err := toMap(fv.Index(k).Interface(), tag)
					if err != nil {
						return nil, err
					}
					s[k] = m
				}
				out[key] = s
			} else {
				out[key] = fv.Interface()
			}
		default:
			if key != "" && !fv.IsZero() {
				out[key] = fv.Interface()
			}
		}
	}
	return out, nil
}
