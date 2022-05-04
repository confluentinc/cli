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

	ver := pversion.NewVersion(version, commit, date, host)

	isTest, err := strconv.ParseBool(isTest)
	if err != nil {
		panic(err)
	}
	cfg.IsTest = isTest

	cli := cmd.NewConfluentCommand(cfg, ver, isTest)

	if !cmd.Execute(cli, cfg, ver, isTest) {
		if isTest {
			bincover.ExitCode = 1
		} else {
			os.Exit(1)
		}
	}
}
