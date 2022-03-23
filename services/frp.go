package services

import (
	_ "github.com/fatedier/frp/assets/frpc"
	frpc "github.com/fatedier/frp/cmd/frpc/sub"
	"github.com/fatedier/golib/crypto"
	"math/rand"
	"time"
)

func runFrpClient() {
	crypto.DefaultSalt = "frp"
	rand.Seed(time.Now().UnixNano())
	// Change program arguments for frpc to parse
	// No need to change it for now
	frpc.Execute()
}
