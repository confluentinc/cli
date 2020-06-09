package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/confluentinc/bincover"
	"github.com/jonboulle/clockwork"
	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/viper"

	"github.com/confluentinc/cli/internal/cmd"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/log"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	"github.com/confluentinc/cli/mock"
)

var (
	// Injected from linker flags like `go build -ldflags "-X main.version=$VERSION" -X ...`
	version    = "v0.0.0"
	commit     = ""
	date       = ""
	host       = ""
	cliName    = "confluent"
	segmentKey = "KDsYPLPBNVB1IPJIN5oqrXnxQT9iKezo"
	isTest     = "false"
)

func main() {
	isTest, err := strconv.ParseBool(isTest)
	if err != nil {
		panic(err)
	}
	viper.AutomaticEnv()

	logger := log.New()

	version := pversion.NewVersion(cliName, version, commit, date, host)

	var analyticsClient analytics.Client
	if !isTest && cliName == "ccloud" {
		segmentClient, _ := segment.NewWithConfig(segmentKey, segment.Config{
			Logger: analytics.NewLogger(logger),
		})

		analyticsClient = analytics.NewAnalyticsClient(cliName, version.Version, segmentClient, clockwork.NewRealClock())
	} else {
		analyticsClient = mock.NewDummyAnalyticsMock()
	}

	cli, err := cmd.NewConfluentCommand(cliName, logger, version, analyticsClient, pauth.NewNetrcHandler(pauth.GetNetrcFilePath(isTest)))
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
			exit(1, analyticsClient, logger)
		}
	}
	err = cli.Execute(cliName, os.Args[1:])
	if err != nil {
		if isTest {
			bincover.ExitCode = 1
			return
		} else {
			exit(1, analyticsClient, logger)
		}
	}
	exit(0, analyticsClient, logger)
}

func exit(exitCode int, analytics analytics.Client, logger *log.Logger) {
	err := analytics.Close()
	if err != nil {
		logger.Debug(err)
	}
	if exitCode == 1 {
		os.Exit(exitCode)
	}
	// no os.Exit(0) because it will shutdown integration test
}
