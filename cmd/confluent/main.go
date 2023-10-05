package main

import (
	"os"
	"strconv"

	"github.com/confluentinc/cli/v3/internal"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	pversion "github.com/confluentinc/cli/v3/pkg/version"
)

// Injected from linker flags like `go build -ldflags "-X main.version=$VERSION" -X ...`
var (
	version        = "0.0.0"
	commit         = ""
	date           = ""
	disableUpdates = "false"
	isTest         = "false"
)

func main() {
	cfg := config.New()

	err := cfg.Load()
	pcmd.CheckErr(cfg.EnableColor, err)

	disableUpdates, err := strconv.ParseBool(disableUpdates)
	pcmd.CheckErr(cfg.EnableColor, err)

	isTest, err := strconv.ParseBool(isTest)
	pcmd.CheckErr(cfg.EnableColor, err)

	cfg.DisableUpdates = disableUpdates
	cfg.IsTest = isTest
	cfg.Version = pversion.NewVersion(version, commit, date)

	cmd := internal.NewConfluentCommand(cfg)

	err = internal.Execute(cmd, os.Args[1:], cfg)
	pcmd.CheckErr(cfg.EnableColor, err)
}
