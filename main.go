package main

import (
	"fmt"
	"os"

	"github.com/confluentinc/cli/command"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/command/config"
	"github.com/hashicorp/go-plugin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/confluentinc/cli/command/auth"
	"github.com/confluentinc/cli/command/connect"
	"github.com/confluentinc/cli/command/kafka"
	"github.com/confluentinc/cli/command/ksql"
	log "github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/metric"
	"github.com/confluentinc/cli/shared"
)

var (
	// Injected from linker flag like `go build -ldflags "-X main.version=$VERSION"`
	version = "0.0.0"

	cli = &cobra.Command{
		Use:   "confluent",
		Short: "Run the Confluent CLI",
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

	cli.Version = version

	// Automatically stop plugins when CLI exits
	cli.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		plugin.CleanupClients()
	}

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

	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		os.Exit(1)
	}
}
