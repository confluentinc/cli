package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"

	"github.com/confluentinc/cli/internal/cmd"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/metric"
	"github.com/confluentinc/cli/internal/pkg/test-integ"
	cliVersion "github.com/confluentinc/cli/internal/pkg/version"
)

var (
	// Injected from linker flags like `go build -ldflags "-X main.version=$VERSION" -X ...`
	version = "v0.0.0"
	commit  = ""
	date    = ""
	host    = ""
	cliName = "confluent"
	isTest  = "false"
)

func main() {
	isTest, err := strconv.ParseBool(isTest)
	if err != nil {
		panic(err)
	}
	viper.AutomaticEnv()

	logger := log.New()

	metricSink := metric.NewSink()

	var cfg *config.Config

	cfg = config.New(&config.Config{
		CLIName:    cliName,
		MetricSink: metricSink,
		Logger:     logger,
	})
	err = cfg.Load()
	if err != nil {
		logger.Errorf("unable to load config: %v", err)
	}

	version := cliVersion.NewVersion(cfg.CLIName, cfg.Name(), cfg.Support(), version, commit, date, host)

	cli, err := cmd.NewConfluentCommand(cliName, cfg, version, logger)
	if err != nil {
		if cli == nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			pcmd.ErrPrintln(cli, err)
		}
		if isTest {
			test_integ.ExitCode = 1
		} else {
			os.Exit(1)
		}
	}
	err = cli.Execute()
	if err != nil {
		if isTest {
			test_integ.ExitCode = 1
		} else {
			os.Exit(1)
		}
	}
}
