package main

import (
	"os"
	"strconv"

	"github.com/confluentinc/bincover"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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
	viper.AutomaticEnv()

	cfg, err := cmd.LoadConfig()
	if err != nil {
		cobra.CheckErr(err)
	}

	isTest, err := strconv.ParseBool(isTest)
	if err != nil {
		panic(err)
	}
	cfg.IsTest = isTest

	version := pversion.NewVersion(version, commit, date, host)

	cli := cmd.New(cfg, isTest, version)

	if err := cli.Execute(os.Args[1:]); err != nil {
		if isTest {
			bincover.ExitCode = 1
		} else {
			os.Exit(1)
		}
	}
}
