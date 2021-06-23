package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"

	"github.com/confluentinc/bincover"
	"github.com/confluentinc/cli/internal/cmd"
	pconfig "github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/config/load"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/metric"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

var (
	// Injected from linker flags like `go build -ldflags "-X main.version=$VERSION" -X ...`
	version = "v0.0.0"
	commit  = ""
	date    = ""
	host    = ""
	isTest  = "false"
)

func main() {
	viper.AutomaticEnv()

	version := pversion.NewVersion(version, commit, date, host)

	isTest, err := strconv.ParseBool(isTest)
	if err != nil {
		panic(err)
	}

	cfg, err := loadV3Config()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, err.Error())
		if isTest {
			bincover.ExitCode = 1
			return
		} else {
			os.Exit(1)
		}
	}

	cli := cmd.NewConfluentCommand(cfg, isTest, version)

	if err := cli.Execute(os.Args[1:]); err != nil {
		if isTest {
			bincover.ExitCode = 1
			return
		} else {
			os.Exit(1)
		}
	}
}

func loadV3Config() (*v3.Config, error) {
	cfg := v3.New(&pconfig.Params{
		Logger:     log.New(),
		MetricSink: metric.NewSink(),
	})

	return load.LoadAndMigrate(cfg)
}
