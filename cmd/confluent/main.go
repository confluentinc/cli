package main

import (
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/internal"
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
	cobra.CheckErr(err)

	disableUpdates, err := strconv.ParseBool(disableUpdates)
	cobra.CheckErr(err)

	isTest, err := strconv.ParseBool(isTest)
	cobra.CheckErr(err)

	cfg.DisableUpdates = disableUpdates
	cfg.IsTest = isTest
	cfg.Version = pversion.NewVersion(version, commit, date)

	cmd := internal.NewConfluentCommand(cfg)

	if err := internal.Execute(cmd, os.Args[1:], cfg); err != nil {
		os.Exit(1)
	}
}
