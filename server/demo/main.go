package main

import (
	"github.com/overtalk/bgo/app"

	_ "github.com/overtalk/bgo/internal/pprof/src"
	_ "github.com/overtalk/bgo/modules/mysql/src"
	_ "github.com/overtalk/bgo/modules/redis/src"
)

func main() { app.Start() }
