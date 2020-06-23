package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/confluentinc/bincover"
	"github.com/spf13/viper"

	"github.com/confluentinc/cli/internal/cmd"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

var (
	// Injected from linker flags like `go build -ldflags "-X main.version=$VERSION" -X ...`
	version    = "v0.0.0"
	commit     = ""
	date       = ""
	host       = ""
	cliName    = "confluent"
	isTest     = "false"
)

func main() {
	isTest, err := strconv.ParseBool(isTest)
	if err != nil {
		panic(err)
	}
	viper.AutomaticEnv()

	version := pversion.NewVersion(cliName, version, commit, date, host)

	cli, err := cmd.NewConfluentCommand(cliName, isTest, version, pauth.NewNetrcHandler(pauth.GetNetrcFilePath(isTest)))
	if err != nil {
		if cli == nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			pcmd.ErrPrintln(cli.Command, err)
		}
		if isTest {
			bincover.ExitCode = 1
			return
		} else {
			exit(1)
		}
	}
	err = cli.Execute(cliName, os.Args[1:])
	if err != nil {
		if isTest {
			bincover.ExitCode = 1
			return
		} else {
			exit(1)
		}
	}
	exit(0)
}


func exit(exitCode int) {
	if exitCode == 1 {
		os.Exit(exitCode)
	}
	// no os.Exit(0) because it will shutdown integration test
}
