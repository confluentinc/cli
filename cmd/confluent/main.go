package main

import (
	"github.com/confluentinc/cli/internal/pkg/output"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/load"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	ppanic "github.com/confluentinc/cli/internal/pkg/panic"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
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
	cfg, err := load.LoadAndMigrate(v1.New())
	cobra.CheckErr(err)

	disableUpdates, err := strconv.ParseBool(disableUpdates)
	cobra.CheckErr(err)

	isTest, err := strconv.ParseBool(isTest)
	cobra.CheckErr(err)

	cfg.DisableUpdates = disableUpdates
	cfg.IsTest = isTest
	cfg.Version = pversion.NewVersion(version, commit, date)

	cmd := pcmd.NewConfluentCommand(cfg)

	if err := pcmd.Execute(cmd, os.Args[1:], cfg); err != nil {
		if _, ok := err.(*ppanic.Panic); ok {
			output.ErrPrintln(err.Error())
		}
		os.Exit(1)
	}
}
