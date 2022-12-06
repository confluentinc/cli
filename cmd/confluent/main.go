package main

import (
	"os"
	"strconv"

	"github.com/confluentinc/bincover"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/load"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

// Injected from linker flags like `go build -ldflags "-X main.version=$VERSION" -X ...`
var (
	version = "v0.0.0"
	commit  = ""
	date    = ""
	isTest  = "false"
)

func main() {
	cfg, err := load.LoadAndMigrate(v1.New())
	cobra.CheckErr(err)

	isTest, err := strconv.ParseBool(isTest)
	cobra.CheckErr(err)

	cfg.IsTest = isTest
	cfg.Version = pversion.NewVersion(version, commit, date)

	cmd := pcmd.NewConfluentCommand(cfg)

	if err := pcmd.Execute(cmd, os.Args[1:], cfg); err != nil {
		if isTest {
			bincover.ExitCode = 1
		} else {
			os.Exit(1)
		}
	}
}
