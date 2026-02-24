//go:build ignore

// generates environment variables.

package main

import (
	"fmt"
	"time"

	"github.com/koho/frpmgr/pkg/version"
)

func main() {
	fmt.Println("VERSION=" + version.Number)
	fmt.Println("BUILD_DATE=" + time.Now().Format(time.DateOnly))
}
