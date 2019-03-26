package main

import (
	"os"

	"github.com/spf13/viper"

	"github.com/confluentinc/cli/internal/cmd"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/metric"
	//connectp "github.com/confluentinc/cli/pkg/connect"
	iconfig "github.com/confluentinc/cli/internal/pkg/config"
	cliVersion "github.com/confluentinc/cli/internal/pkg/version"
)

var (
	// Injected from linker flags like `go build -ldflags "-X main.version=$VERSION" -X ...`
	version = "v0.0.0"
	commit  = ""
	date    = ""
	host    = ""
)

func main() {
	viper.AutomaticEnv()

	logger := log.New()

	metricSink := metric.NewSink()

	var cfg *iconfig.Config
	{
		cfg = iconfig.NewConfig(&iconfig.Config{
			MetricSink: metricSink,
			Logger:     logger,
		})
		err := cfg.Load()
		if err != nil && err != iconfig.ErrNoConfig {
			logger.Errorf("unable to load config: %v", err)
		}
	}

	version := cliVersion.NewVersion(version, commit, date, host)

	cli := cmd.NewConfluentCommand(cfg, version, logger)
	err := cli.Execute()
	if err != nil {
		os.Exit(1)
	}
}
