package main

import (
	"os"
	"strconv"

	"github.com/confluentinc/bincover"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd"
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
	cfg, err := cmd.LoadConfig()
	cobra.CheckErr(err)

	isTest, err := strconv.ParseBool(isTest)
	if err != nil {
		panic(err)
	}
	cfg.IsTest = isTest

	version := pversion.NewVersion(version, commit, date, host)

	cli := cmd.NewConfluentCommand(cfg, isTest, version)

	if !cmd.Execute(cli, cfg) {
		if isTest {
			bincover.ExitCode = 1
		} else {
			os.Exit(1)
		}
	}
}
