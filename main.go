package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/hashicorp/go-plugin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/confluentinc/cli/command"
	"github.com/confluentinc/cli/command/auth"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/command/config"
	"github.com/confluentinc/cli/command/connect"
	"github.com/confluentinc/cli/command/kafka"
	"github.com/confluentinc/cli/command/ksql"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/metric"
	"github.com/confluentinc/cli/shared"
	cliVersion "github.com/confluentinc/cli/version"
)

var (
	// Injected from linker flags like `go build -ldflags "-X main.version=$VERSION" -X ...`
	version = "v0.0.0"
	commit  = ""
	date    = ""
	host    = ""

	cli = &cobra.Command{
		Use:   os.Args[0],
		Short: "Welcome to the Confluent Cloud CLI",
	}
)

func main() {
	viper.AutomaticEnv()

	var logger *log.Logger
	{
		logger = log.New()
		logger.Out = os.Stdout
		logger.Log("msg", "hello")
		defer logger.Log("msg", "goodbye")

		if viper.GetString("log_level") != "" {
			level, err := logrus.ParseLevel(viper.GetString("log_level"))
			check(err)
			logger.SetLevel(level)
			logger.Log("msg", "set log level", "level", level.String())
		}
	}

	var metricSink shared.MetricSink
	{
		metricSink = metric.NewSink()
	}

	var cfg *shared.Config
	{
		cfg = shared.NewConfig(&shared.Config{
			MetricSink: metricSink,
			Logger:     logger,
		})
		err := cfg.Load()
		if err != nil && err != shared.ErrNoConfig {
			logger.WithError(err).Errorf("unable to load cfg")
		}
	}

	prompt := command.NewTerminalPrompt(os.Stdin)

	userAgent := fmt.Sprintf("Confluent/1.0 ccloud/%s (%s/%s)", version, runtime.GOOS, runtime.GOARCH)
	version := cliVersion.NewVersion(version, commit, date, host, userAgent)

	cli.Version = version.Version
	cli.AddCommand(common.NewVersionCmd(version, prompt))

	cli.AddCommand(config.New(cfg))

	cli.AddCommand(common.NewCompletionCmd(cli, prompt))

	cli.AddCommand(auth.New(cfg)...)

	conn, err := kafka.New(cfg)
	if err != nil {
		logger.Log("msg", err)
	} else {
		cli.AddCommand(conn)
	}

	conn, err = connect.New(cfg)
	if err != nil {
		logger.Log("msg", err)
	} else {
		cli.AddCommand(conn)
	}

	conn, err = ksql.New(cfg)
	if err != nil {
		logger.Log("msg", err)
	} else {
		cli.AddCommand(conn)
	}

	check(cli.Execute())

	plugin.CleanupClients()
	os.Exit(0)
}

func check(err error) {
	if err != nil {
		plugin.CleanupClients()
		os.Exit(1)
	}
}
