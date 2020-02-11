package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/confluentinc/bincover"
	"github.com/jonboulle/clockwork"
	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/confluentinc/cli/internal/cmd"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/config/load"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/metric"
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

	metricSink := metric.NewSink()

	var cfg *v2.Config

	params := &config.Params{
		CLIName:    cliName,
		MetricSink: metricSink,
		Logger:     logger,
	}
	cfg = v2.New(params)
	cfg, err = load.LoadAndMigrate(cfg)
	if err != nil {
		stubCmd := &cobra.Command{}
		err = errors.HandleCommon(err, stubCmd)
		errFmt := "unable to load config: %v\n"
		logger.Debug(errFmt, err)
		fmt.Fprintf(os.Stderr, errFmt, err)
		if isTest {
			bincover.ExitCode = 1
			return
		} else {
			os.Exit(1)
		}
	}

	version := pversion.NewVersion(cfg.CLIName, cfg.Name(), cfg.Support(), version, commit, date, host)

	var analyticsClient analytics.Client
	if !isTest && cfg.CLIName == "ccloud" {
		segmentClient, _ := segment.NewWithConfig(segmentKey, segment.Config{
			Logger: analytics.NewLogger(logger),
		})

		analyticsClient = analytics.NewAnalyticsClient(cfg.CLIName, cfg, version.Version, segmentClient, clockwork.NewRealClock())
	} else {
		analyticsClient = mock.NewDummyAnalyticsMock()
	}

	cli, err := cmd.NewConfluentCommand(cliName, cfg, logger, version, analyticsClient)
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
	err = cli.Execute(os.Args[1:])
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
