package cmd

import (
	"github.com/confluentinc/cli/internal/cmd/auditlog"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"net/http"
	"os"
	"runtime"

	"github.com/DABH/go-basher"
	"github.com/jonboulle/clockwork"
	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/apikey"
	"github.com/confluentinc/cli/internal/cmd/auth"
	"github.com/confluentinc/cli/internal/cmd/cluster"
	"github.com/confluentinc/cli/internal/cmd/completion"
	"github.com/confluentinc/cli/internal/cmd/config"
	"github.com/confluentinc/cli/internal/cmd/connect"
	"github.com/confluentinc/cli/internal/cmd/connector"
	connector_catalog "github.com/confluentinc/cli/internal/cmd/connector-catalog"
	"github.com/confluentinc/cli/internal/cmd/environment"
	"github.com/confluentinc/cli/internal/cmd/feedback"
	"github.com/confluentinc/cli/internal/cmd/iam"
	initcontext "github.com/confluentinc/cli/internal/cmd/init-context"
	"github.com/confluentinc/cli/internal/cmd/kafka"
	"github.com/confluentinc/cli/internal/cmd/ksql"
	"github.com/confluentinc/cli/internal/cmd/local"
	ps1 "github.com/confluentinc/cli/internal/cmd/prompt"
	schema_registry "github.com/confluentinc/cli/internal/cmd/schema-registry"
	"github.com/confluentinc/cli/internal/cmd/secret"
	service_account "github.com/confluentinc/cli/internal/cmd/service-account"
	"github.com/confluentinc/cli/internal/cmd/update"
	"github.com/confluentinc/cli/internal/cmd/version"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	pconfig "github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/config/load"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	pfeedback "github.com/confluentinc/cli/internal/pkg/feedback"
	"github.com/confluentinc/cli/internal/pkg/help"
	"github.com/confluentinc/cli/internal/pkg/io"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/metric"
	pps1 "github.com/confluentinc/cli/internal/pkg/ps1"
	secrets "github.com/confluentinc/cli/internal/pkg/secret"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	"github.com/confluentinc/cli/mock"
)

var segmentKey = "KDsYPLPBNVB1IPJIN5oqrXnxQT9iKezo"

type Command struct {
	*cobra.Command
	// @VisibleForTesting
	Analytics analytics.Client
	logger    *log.Logger
}

func NewConfluentCommand(cliName string, isTest bool, ver *pversion.Version, netrcHandler *pauth.NetrcHandler) (*Command, error) {
	logger := log.New()
	cfg, configLoadingErr := loadConfig(cliName, logger)
	if cfg != nil {
		cfg.Logger = logger
	}
	analyticsClient := getAnalyticsClient(isTest, cliName, cfg, ver.Version, logger)
	cli := &cobra.Command{
		Use:               cliName,
		Version:           ver.Version,
		DisableAutoGenTag: true,
	}
	cli.SetUsageFunc(func(cmd *cobra.Command) error {
		return help.ResolveReST(cmd.UsageTemplate(), cmd)
	})
	cli.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		_ = help.ResolveReST(cmd.HelpTemplate(), cmd)
	})
	if cliName == "ccloud" {
		cli.Short = "Confluent Cloud CLI."
		cli.Long = "Manage your Confluent Cloud."
	} else {
		cli.Short = "Confluent CLI."
		cli.Long = "Manage your Confluent Platform."
	}
	cli.PersistentFlags().CountP("verbose", "v",
		"Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).")

	prompt := pcmd.NewPrompt(os.Stdin)

	disableUpdateCheck := cfg != nil && (cfg.DisableUpdates || cfg.DisableUpdateCheck)
	updateClient, err := update.NewClient(cliName, disableUpdateCheck,logger)
	if err != nil {
		return nil, err
	}

	resolver := &pcmd.FlagResolverImpl{Prompt: prompt, Out: os.Stdout}
	prerunner := &pcmd.PreRun{
		Config: cfg,
		ConfigLoadingError: configLoadingErr,
		UpdateClient:       updateClient,
		CLIName:            cliName,
		Logger:             logger,
		Clock:              clockwork.NewRealClock(),
		FlagResolver:       resolver,
		Version:            ver,
		Analytics:          analyticsClient,
		UpdateTokenHandler: pauth.NewUpdateTokenHandler(netrcHandler),
	}
	command := &Command{Command: cli, Analytics: analyticsClient, logger: logger}

	cli.Version = ver.Version
	cli.AddCommand(version.NewVersionCmd(prerunner, ver))

	cli.AddCommand(completion.NewCompletionCmd(cli, cliName))

	if cfg == nil || !cfg.DisableUpdates {
		cli.AddCommand(update.New(cliName, logger, ver, prompt, updateClient, analyticsClient))
	}

	cli.AddCommand(auth.New(cliName, prerunner, logger, ver.UserAgent, analyticsClient, netrcHandler)...)
	isAPILogin := isAPIKeyCredential(cfg)
	if cliName == "ccloud" {
		cmd := kafka.New(isAPILogin, cliName, prerunner, logger.Named("kafka"), ver.ClientID)
		cli.AddCommand(cmd)
		cli.AddCommand(feedback.NewFeedbackCmd(cliName, prerunner, analyticsClient))
		cli.AddCommand(initcontext.New(prerunner, prompt, resolver, analyticsClient))
		cli.AddCommand(config.New(prerunner, analyticsClient))
		if isAPILogin {
			return command, nil
		}
		cli.AddCommand(ps1.NewPromptCmd(cliName, prerunner, &pps1.Prompt{}, logger))
		cli.AddCommand(environment.New(cliName, prerunner))
		cli.AddCommand(service_account.New(prerunner))
		// Keystore exposed so tests can pass mocks.
		cli.AddCommand(apikey.New(prerunner, nil, resolver))

		// Schema Registry
		// If srClient is nil, the function will look it up after prerunner verifies authentication. Exposed so tests can pass mocks
		cli.AddCommand(schema_registry.New(cliName, prerunner, nil, logger))
		cli.AddCommand(ksql.New(cliName, prerunner))
		cli.AddCommand(connector.New(cliName, prerunner))
		cli.AddCommand(connector_catalog.New(cliName, prerunner))
		//conn = connect.New(prerunner, cfg, connects.New(client, logger))
		//conn.Hidden = true // The connect feature isn't finished yet, so let's hide it
		//cli.AddCommand(conn)
	} else if cliName == "confluent" {
		cli.AddCommand(iam.New(cliName, prerunner))
		// Kafka Command
		isAPILogin := isAPIKeyCredential(cfg)
		cmd := kafka.New(isAPILogin, cliName, prerunner, logger.Named("kafka"), ver.ClientID)
		cli.AddCommand(cmd)
		sr := schema_registry.New(cliName, prerunner, nil, logger)
		cli.AddCommand(sr)
		cli.AddCommand(ksql.New(cliName, prerunner))
		cli.AddCommand(connect.New(prerunner))

		metaClient := cluster.NewScopedIdService(&http.Client{}, ver.UserAgent, logger)
		cli.AddCommand(cluster.New(prerunner, metaClient))

		if runtime.GOOS != "windows" {
			bash, err := basher.NewContext("/bin/bash", false)
			if err != nil {
				return nil, err
			}
			shellRunner := &local.BashShellRunner{BasherContext: bash}
			cli.AddCommand(local.New(cli, prerunner, shellRunner, logger, &io.RealFileSystem{}))
		}

		command := local.NewCommand(prerunner)
		command.Hidden = true // WIP
		cli.AddCommand(command)

		cli.AddCommand(secret.New(prompt, resolver, secrets.NewPasswordProtectionPlugin(logger)))

		cli.AddCommand(auditlog.New(prerunner))
	}
	return command, nil
}

func getAnalyticsClient(isTest bool, cliName string, cfg *v3.Config, cliVersion string, logger *log.Logger) analytics.Client {
	if cliName == "confluent" || isTest {
		return mock.NewDummyAnalyticsMock()
	}
	segmentClient, _ := segment.NewWithConfig(segmentKey, segment.Config{
		Logger: analytics.NewLogger(logger),
	})
	return analytics.NewAnalyticsClient(cliName, cfg, cliVersion, segmentClient, clockwork.NewRealClock())
}

func isAPIKeyCredential(cfg *v3.Config) bool {
	if cfg == nil {
		return false
	}
	currCtx := cfg.Context()
	if currCtx != nil && currCtx.Credential != nil && currCtx.Credential.CredentialType == v2.APIKey {
		return true
	}
	return false
}

func (c *Command) Execute(cliName string, args []string) error {
	c.Analytics.SetStartTime()
	c.Command.SetArgs(args)

	err := c.Command.Execute()
	c.sendAndFlushAnalytics(args, err)
	pfeedback.HandleFeedbackNudge(cliName, args)
	return err
}

func (c *Command) sendAndFlushAnalytics(args []string, err error) {
	analyticsError := c.Analytics.SendCommandAnalytics(c.Command, args, err)
	if analyticsError != nil {
		c.logger.Debugf("segment analytics sending event failed: %s\n", analyticsError.Error())
	}
	err = c.Analytics.Close()
	if err != nil {
		c.logger.Debug(err)
	}
}

func loadConfig(cliName string, logger *log.Logger) (*v3.Config, error) {
	metricSink := metric.NewSink()
	params := &pconfig.Params{
		CLIName:    cliName,
		MetricSink: metricSink,
		Logger:     logger,
	}
	cfg := v3.New(params)
	cfg, err := load.LoadAndMigrate(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, err
}
