package main

import (
	"fmt"
	"os"
	"path/filepath"

	//"github.com/confluentinc/cli/command"
	//"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/command/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/confluentinc/cli/command/auth"
	log "github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/metric"
	"github.com/confluentinc/cli/shared"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	"os/exec"
	"strings"
)

var plugins []*cobra.Command
var cachedir string

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

	cli.Version = version
	// Only show commands if we are logged in and have plugins loaded
	if cfg.CheckLogin() == nil && len(plugins) > 0 {
		cli.AddCommand(plugins...)
	} else {
		cli.AddCommand(config.New(cfg))
		cli.AddCommand(auth.New(cfg)...)
	}

	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}

func runner(c *cobra.Command, args []string) {
	cmd := exec.Command(filepath.Join(cachedir, fmt.Sprintf("confluent-%s-plugin", c.Name())), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	check(cmd.Run())
}

func deferHelp(cmd *cobra.Command, args []string) {
	runner(cmd, args[1:])
}

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		os.Exit(1)
	}
}

func initCache() {
	var err error

	cachedir, err = os.UserCacheDir()
	if err != nil {
		cachedir, err = homedir.Expand("~")
		check(err)
	}

	cachedir = filepath.Join(cachedir, "confluent")

	if _, err := os.Stat(cachedir); os.IsNotExist(err) {
		check(os.Mkdir(cachedir, 0700))
	}
}

// Example, not actually how it should work
func init() {
	// remove redundant help command
	cli.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})

	initCache()
	files, err := ioutil.ReadDir(cachedir)
	check(err)

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "confluent-") {
			tmp := strings.Split(file.Name(), "-")
			if len(tmp) < 3 {
				continue
			}
			cmd := &cobra.Command{
				Use:   tmp[1],
				Short: "Run Confluent " + strings.Title(tmp[1]) + " CLI",
				Run:   runner,
			}
			cmd.SetHelpFunc(deferHelp)
			plugins = append(plugins, cmd)
		}
	}
}
