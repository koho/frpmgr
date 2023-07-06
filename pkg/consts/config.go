package consts

import "github.com/fatedier/frp/pkg/config"

const RangePrefix = "range:"

// Protocols
const (
	ProtoTCP       = "tcp"
	ProtoKCP       = "kcp"
	ProtoQUIC      = "quic"
	ProtoWebsocket = "websocket"
	ProtoWSS       = "wss"
)

var Protocols = []string{ProtoTCP, ProtoKCP, ProtoQUIC, ProtoWebsocket, ProtoWSS}

// Proxy types
const (
	ProxyTypeTCP    = "tcp"
	ProxyTypeUDP    = "udp"
	ProxyTypeXTCP   = "xtcp"
	ProxyTypeSTCP   = "stcp"
	ProxyTypeSUDP   = "sudp"
	ProxyTypeHTTP   = "http"
	ProxyTypeHTTPS  = "https"
	ProxyTypeTCPMUX = "tcpmux"
)

var ProxyTypes = []string{
	ProxyTypeTCP, ProxyTypeUDP, ProxyTypeXTCP, ProxyTypeSTCP,
	ProxyTypeSUDP, ProxyTypeHTTP, ProxyTypeHTTPS, ProxyTypeTCPMUX,
}

// Plugin types
const (
	PluginHttpProxy   = "http_proxy"
	PluginSocks5      = "socks5"
	PluginStaticFile  = "static_file"
	PluginHttps2Http  = "https2http"
	PluginHttps2Https = "https2https"
	PluginHttp2Https  = "http2https"
	PluginUnixDomain  = "unix_domain_socket"
)

var PluginTypes = []string{
	PluginHttpProxy, PluginSocks5, PluginStaticFile,
	PluginHttps2Http, PluginHttps2Https, PluginHttp2Https,
	PluginUnixDomain,
}

// Auth methods
const (
	AuthToken = "token"
	AuthOIDC  = "oidc"
)

// Delete methods
const (
	DeleteAbsolute = "absolute"
	DeleteRelative = "relative"
)

// TCP multiplexer
const (
	HTTPConnectTCPMultiplexer = "httpconnect"
)

// Bandwidth
var (
	Bandwidth     = []string{"MB", "KB"}
	BandwidthMode = []string{
		config.BandwidthLimitModeClient,
		config.BandwidthLimitModeServer,
	}
)

// Log level
const (
	LogLevelTrace = "trace"
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

var LogLevels = []string{LogLevelTrace, LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError}
