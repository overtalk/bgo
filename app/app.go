package app

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/overtalk/bgo/core"
)

var (
	// some info about git which should be set in `go build`
	commit  string
	branch  string
	version = "no-version"

	showVersion bool
	confPath    string
)

func parseFlags() error {
	flag.StringVar(&confPath, "p", "", "config dir path")
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.Parse()
	return nil
}

func printVersion() {
	fmt.Printf("Version : %s \nBranch : %s \nCommitID : %s\n", version, branch, commit)
}

func Start() {
	if err := parseFlags(); err != nil {
		log.Fatal(err)
	}

	if showVersion {
		printVersion()
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	core.GetCore().SetNotifyChan(sigChan)
	core.GetCore().SetConfigPath(confPath)

	if err := core.GetCore().Start(); err != nil {
		log.Fatal(err)
	}
	defer core.GetCore().Stop()

	core.GetCore().Ticker()

	<-sigChan
}
