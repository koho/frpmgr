package services

import (
	_ "github.com/fatedier/frp/assets/frpc"
	frpc "github.com/fatedier/frp/cmd/frpc/sub"
	"github.com/fatedier/frp/pkg/config"
)

func runFrpClient() {
	// Change program arguments for frpc to parse
	// No need to change it for now
	frpc.Execute()
}

// VerifyClientConfig validates the frp client config file
func VerifyClientConfig(path string) (err error) {
	_, _, _, err = config.ParseClientConfig(path)
	return
}

// VerifyClientProxy validates the frp proxy
func VerifyClientProxy(source []byte) (err error) {
	_, _, err = config.LoadAllProxyConfsFromIni("", source, nil)
	return
}
