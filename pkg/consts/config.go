package consts

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
