package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/jonboulle/clockwork"
	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/apikey"
	"github.com/confluentinc/cli/internal/cmd/auditlog"
	"github.com/confluentinc/cli/internal/cmd/auth"
	"github.com/confluentinc/cli/internal/cmd/cluster"
	"github.com/confluentinc/cli/internal/cmd/completion"
	"github.com/confluentinc/cli/internal/cmd/config"
	"github.com/confluentinc/cli/internal/cmd/connect"
	"github.com/confluentinc/cli/internal/cmd/connector"
	connectorcatalog "github.com/confluentinc/cli/internal/cmd/connector-catalog"
	"github.com/confluentinc/cli/internal/cmd/environment"
	"github.com/confluentinc/cli/internal/cmd/feedback"
	"github.com/confluentinc/cli/internal/cmd/iam"
	initcontext "github.com/confluentinc/cli/internal/cmd/init-context"
	"github.com/confluentinc/cli/internal/cmd/kafka"
	"github.com/confluentinc/cli/internal/cmd/ksql"
	"github.com/confluentinc/cli/internal/cmd/local"
	ps1 "github.com/confluentinc/cli/internal/cmd/prompt"
	schemaregistry "github.com/confluentinc/cli/internal/cmd/schema-registry"
	"github.com/confluentinc/cli/internal/cmd/secret"
	serviceaccount "github.com/confluentinc/cli/internal/cmd/service-account"
	"github.com/confluentinc/cli/internal/cmd/update"
	"github.com/confluentinc/cli/internal/cmd/version"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	pconfig "github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/config/load"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	pfeedback "github.com/confluentinc/cli/internal/pkg/feedback"
	"github.com/confluentinc/cli/internal/pkg/help"
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

	cli.PersistentFlags().BoolP("help", "h", false, "Show help for this command.")
	cli.PersistentFlags().CountP("verbose", "v", "Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).")
	cli.Flags().Bool("version", false, fmt.Sprintf("Show version of %s.", cliName))

	disableUpdateCheck := cfg != nil && (cfg.DisableUpdates || cfg.DisableUpdateCheck)
	updateClient, err := update.NewClient(cliName, disableUpdateCheck, logger)
	if err != nil {
		return nil, err
	}

	resolver := &pcmd.FlagResolverImpl{Prompt: pcmd.NewPrompt(os.Stdin), Out: os.Stdout}
	prerunner := &pcmd.PreRun{
		Config:             cfg,
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
	cli.AddCommand(version.New(cliName, prerunner, ver))

	cli.AddCommand(completion.New(cli, cliName))

	if cfg == nil || !cfg.DisableUpdates {
		cli.AddCommand(update.New(cliName, logger, ver, updateClient, analyticsClient))
	}

	cli.AddCommand(auth.New(cliName, prerunner, logger, ver.UserAgent, analyticsClient, netrcHandler)...)
	isAPILogin := isAPIKeyCredential(cfg)
	if cliName == "ccloud" {
		cli.AddCommand(config.New(prerunner, analyticsClient))
		cli.AddCommand(feedback.New(cliName, prerunner, analyticsClient))
		cli.AddCommand(initcontext.New(prerunner, resolver, analyticsClient))
		cli.AddCommand(kafka.New(isAPILogin, cliName, prerunner, logger.Named("kafka"), ver.ClientID))
		if isAPIKeyCredential(cfg) {
			return command, nil
		}
		cli.AddCommand(apikey.New(prerunner, nil, resolver)) // Exposed for testing
		cli.AddCommand(connector.New(cliName, prerunner))
		cli.AddCommand(connectorcatalog.New(cliName, prerunner))
		cli.AddCommand(environment.New(cliName, prerunner))
		cli.AddCommand(ksql.New(cliName, prerunner))
		cli.AddCommand(ps1.New(cliName, prerunner, &pps1.Prompt{}, logger))
		cli.AddCommand(schemaregistry.New(cliName, prerunner, nil, logger)) // Exposed for testing
		cli.AddCommand(serviceaccount.New(prerunner))
		if os.Getenv("XX_CCLOUD_RBAC") != "" {
			cli.AddCommand(iam.New(cliName, prerunner))
		}
	} else if cliName == "confluent" {
		cli.AddCommand(auditlog.New(prerunner))
		cli.AddCommand(cluster.New(prerunner, cluster.NewScopedIdService(&http.Client{}, ver.UserAgent, logger)))
		cli.AddCommand(connect.New(prerunner))
		cli.AddCommand(iam.New(cliName, prerunner))
		cli.AddCommand(kafka.New(isAPIKeyCredential(cfg), cliName, prerunner, logger.Named("kafka"), ver.ClientID))
		cli.AddCommand(ksql.New(cliName, prerunner))
		cli.AddCommand(local.New(prerunner))
		cli.AddCommand(schemaregistry.New(cliName, prerunner, nil, logger))
		cli.AddCommand(secret.New(resolver, secrets.NewPasswordProtectionPlugin(logger)))
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
	errors.DisplaySuggestionsMessage(err, os.Stderr)
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
