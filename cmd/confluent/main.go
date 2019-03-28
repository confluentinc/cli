package main

import (
	"os"

	"github.com/spf13/viper"

	"github.com/confluentinc/cli/internal/cmd"
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
	cliName = ""
)

func main() {
	viper.AutomaticEnv()

	logger := log.New()

	metricSink := metric.NewSink()

	var cfg *config.Config
	{
		cfg = config.New(&config.Config{
			MetricSink: metricSink,
			Logger:     logger,
		})
		err := cfg.Load()
		if err != nil && err != config.ErrNoConfig {
			logger.Errorf("unable to load config: %v", err)
		}
	}

	version := cliVersion.NewVersion(version, commit, date, host)

	cli := cmd.NewConfluentCommand(cfg, version, logger, cliName)
	err := cli.Execute()
	if err != nil {
		os.Exit(1)
	}
}
