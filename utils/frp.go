package utils

import (
	frpc "github.com/fatedier/frp/cmd/frpc/sub"
	"github.com/fatedier/golib/crypto"
	"math/rand"
	"os"
	"time"
)

func RunFrpClient(c string) {
	crypto.DefaultSalt = "frp"
	rand.Seed(time.Now().UnixNano())
	err := frpc.RunClient(c)
	if err != nil {
		os.Exit(1)
	}
}
