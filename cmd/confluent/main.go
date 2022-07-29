package main

import (
	"os"
	"strconv"

	"github.com/confluentinc/bincover"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/load"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

// Injected from linker flags like `go build -ldflags "-X main.version=$VERSION" -X ...`
var (
	version = "v0.0.0"
	commit  = ""
	date    = ""
	host    = ""
	isTest  = "false"
)

func main() {
	cfg, err := load.LoadAndMigrate(v1.New())
	cobra.CheckErr(err)

	ver := pversion.NewVersion(version, commit, date, host)

	isTest, err := strconv.ParseBool(isTest)
	cobra.CheckErr(err)
	cfg.IsTest = isTest

	cli := cmd.NewConfluentCommand(cfg, ver, isTest)

	if err := cmd.Execute(cli, os.Args[1:], cfg, ver, isTest); err != nil {
		if isTest {
			bincover.ExitCode = 1
		} else {
			os.Exit(1)
		}
	}
}
