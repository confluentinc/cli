package main

import (
	"fmt"
	"os"

	segment "github.com/segmentio/analytics-go"
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
	segmentKey = "KDsYPLPBNVB1IPJIN5oqrXnxQT9iKezo"
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

	segmentClient := segment.New(segmentKey)
	analyticsClient := analytics.NewAnalyticsClient(cfg, segmentClient)

	cli, err := cmd.NewConfluentCommand(cliName, cfg, version, logger, analyticsClient)
	if err != nil {
		if cli == nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			pcmd.ErrPrintln(cli.Command, err)
		}
		os.Exit(1)
	}

	err = cli.Execute()
	if err != nil {
		os.Exit(1)
	}
}
