package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/confluentinc/cli/internal/cmd"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/metric"
	cliVersion "github.com/confluentinc/cli/internal/pkg/version"
)

var (
	// Injected from linker flags like `go build -ldflags "-X main.version=$VERSION" -X ...`
	version = "v0.0.0"
	commit  = ""
	date    = ""
	host    = ""
	cliName = "confluent"
)

func main() {
	viper.AutomaticEnv()

	logger := log.New()

	metricSink := metric.NewSink()

	var cfg *config.Config
	{
		cfg = config.New(&config.Config{
			CLIName:    cliName,
			MetricSink: metricSink,
			Logger:     logger,
		})
		err := cfg.Load()
		if err != nil {
			logger.Errorf("unable to load config: %v", err)
		}
	}

	version := cliVersion.NewVersion(cfg.CLIName, cfg.Name(), cfg.Support(), version, commit, date, host)

	analyticsObject := analytics.NewAnalyticsObject(cfg)
	cli, err := cmd.NewConfluentCommand(cliName, cfg, version, logger)
	if err != nil {
		if cli == nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			pcmd.ErrPrintln(cli, err)
		}
		os.Exit(1)
	}

	analyticsError := analyticsObject.InitializePreRuns(cli)
	if analyticsError != nil {
		logger.Debug("segment analytics set up failed: %s\n", analyticsError.Error())
	}

	err = cli.Execute()
	if err != nil {
		if analyticsError == nil {
			analyticsError = analyticsObject.FlushCommandFailed(err)
			if analyticsError != nil {
				logger.Debug("segment analytics flushing failed: %s\n", analyticsError.Error())
			}
		}
		os.Exit(1)
	}

	if analyticsError == nil {
		analyticsError = analyticsObject.FlushCommandSucceeded()
		if analyticsError != nil {
			logger.Debug("segment analytics flushing failed: %s\n", analyticsError.Error())
		}
	}
}
