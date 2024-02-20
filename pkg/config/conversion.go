package config

import (
	"strconv"
	"strings"

	"github.com/fatedier/frp/pkg/config/v1"
	"github.com/samber/lo"
)

func ClientCommonFromV1(c v1.ClientCommonConfig) (r ClientCommon) {
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
	if c.ServerPort != 0 {
		r.ServerPort = strconv.Itoa(c.ServerPort)
	}
	r.NatHoleSTUNServer = c.NatHoleSTUNServer
	r.DNSServer = c.DNSServer
	if c.LoginFailExit == nil || *c.LoginFailExit {
		r.LoginFailExit = true
	}
	r.Start = c.Start

	// Log
	r.LogFile = c.Log.To
	r.LogLevel = c.Log.Level
	r.LogMaxDays = uint(c.Log.MaxDays)

	// Admin
	r.AdminAddr = c.WebServer.Addr
	if c.WebServer.Port != 0 {
		r.AdminPort = strconv.Itoa(c.WebServer.Port)
	}
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
	r.PoolCount = uint(c.Transport.PoolCount)
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

func ClientProxyFromV1(pxyCfg v1.TypedProxyConfig) *Proxy {
	var r Proxy
	clientProxyBaseFromV1(pxyCfg.GetBaseConfig(), &r)
	switch v := pxyCfg.ProxyConfigurer.(type) {
	case *v1.TCPProxyConfig:
		r.RemotePort = strconv.Itoa(v.RemotePort)
	case *v1.UDPProxyConfig:
		r.RemotePort = strconv.Itoa(v.RemotePort)
	case *v1.HTTPProxyConfig:
		r.SubDomain = v.SubDomain
		r.CustomDomains = strings.Join(v.CustomDomains, ",")
		r.Locations = strings.Join(v.Locations, ",")
		r.HTTPUser = v.HTTPUser
		r.HTTPPwd = v.HTTPPassword
		r.HostHeaderRewrite = v.HostHeaderRewrite
		r.RouteByHTTPUser = v.RouteByHTTPUser
		r.Headers = v.RequestHeaders.Set
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
	out.LocalIP = c.LocalIP
	if c.LocalPort != 0 {
		out.LocalPort = strconv.Itoa(c.LocalPort)
	}

	out.Metas = c.Metadatas
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
		out.HostHeaderRewrite = v.HostHeaderRewrite
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
	if c.BindPort != 0 {
		out.BindPort = strconv.Itoa(c.BindPort)
	}
}
