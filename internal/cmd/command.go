package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/jonboulle/clockwork"
	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/admin"
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
	"github.com/confluentinc/cli/internal/cmd/price"
	ps1 "github.com/confluentinc/cli/internal/cmd/prompt"
	schemaregistry "github.com/confluentinc/cli/internal/cmd/schema-registry"
	"github.com/confluentinc/cli/internal/cmd/secret"
	serviceaccount "github.com/confluentinc/cli/internal/cmd/service-account"
	"github.com/confluentinc/cli/internal/cmd/shell"
	"github.com/confluentinc/cli/internal/cmd/signup"
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
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/help"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/metric"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	pps1 "github.com/confluentinc/cli/internal/pkg/ps1"
	secrets "github.com/confluentinc/cli/internal/pkg/secret"
	"github.com/confluentinc/cli/internal/pkg/shell/completer"
	keys "github.com/confluentinc/cli/internal/pkg/third-party-keys"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	"github.com/confluentinc/cli/mock"
)

type Command struct {
	*cobra.Command
	// @VisibleForTesting
	Analytics analytics.Client
	logger    *log.Logger
}

func NewConfluentCommand(cliName string, isTest bool, ver *pversion.Version, netrcHandler netrc.NetrcHandler) (*Command, error) {
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
	cli.Flags().Bool("version", false, fmt.Sprintf("Show version of the %s.", pversion.GetFullCLIName(cliName)))

	disableUpdateCheck := cfg != nil && (cfg.DisableUpdates || cfg.DisableUpdateCheck)
	updateClient, err := update.NewClient(cliName, disableUpdateCheck, logger)
	if err != nil {
		return nil, err
	}

	authTokenHandler := pauth.NewAuthTokenHandler(logger)
	loginCredentialsManager := pauth.NewLoginCredentialsManager(netrcHandler, form.NewPrompt(os.Stdin), logger)
	resolver := &pcmd.FlagResolverImpl{Prompt: form.NewPrompt(os.Stdin), Out: os.Stdout}
	jwtValidator := pcmd.NewJWTValidator(logger)
	prerunner := &pcmd.PreRun{
		Config:                  cfg,
		ConfigLoadingError:      configLoadingErr,
		UpdateClient:            updateClient,
		CLIName:                 cliName,
		Logger:                  logger,
		FlagResolver:            resolver,
		Version:                 ver,
		Analytics:               analyticsClient,
		LoginCredentialsManager: loginCredentialsManager,
		AuthTokenHandler:        authTokenHandler,
		JWTValidator:            jwtValidator,
	}
	command := &Command{Command: cli, Analytics: analyticsClient, logger: logger}
	shellCompleter := completer.NewShellCompleter(cli)
	serverCompleter := shellCompleter.ServerSideCompleter

	cli.Version = ver.Version
	cli.AddCommand(version.New(cliName, prerunner, ver))

	cli.AddCommand(completion.New(cli, cliName))

	if cfg == nil || !cfg.DisableUpdates {
		cli.AddCommand(update.New(cliName, logger, ver, updateClient, analyticsClient))
	}

	cli.AddCommand(auth.New(cliName, prerunner, logger, ver.UserAgent, analyticsClient, netrcHandler, loginCredentialsManager, authTokenHandler)...)
	isAPILogin := isAPIKeyCredential(cfg)
	cli.AddCommand(config.New(cliName, prerunner, analyticsClient))
	if cliName == "ccloud" {
		cli.AddCommand(admin.New(prerunner, isTest))
		cli.AddCommand(feedback.New(cliName, prerunner, analyticsClient))
		cli.AddCommand(initcontext.New(prerunner, resolver, analyticsClient))
		cli.AddCommand(kafka.New(isAPILogin, cliName, prerunner, logger.Named("kafka"), ver.ClientID, serverCompleter))
		if isAPIKeyCredential(cfg) {
			return command, nil
		}
		apiKeyCmd := apikey.New(prerunner, nil, resolver)
		serverCompleter.AddCommand(apiKeyCmd)
		cli.AddCommand(apiKeyCmd.Command)

		connectorCmd := connector.New(cliName, prerunner)
		serverCompleter.AddCommand(connectorCmd)
		cli.AddCommand(connectorCmd.Command)
		connectorCatalogCmd := connectorcatalog.New(cliName, prerunner)
		serverCompleter.AddCommand(connectorCatalogCmd)
		cli.AddCommand(connectorCatalogCmd.Command)
		envCmd := environment.New(cliName, prerunner)
		serverCompleter.AddCommand(envCmd)
		cli.AddCommand(envCmd.Command)
		cli.AddCommand(ksql.New(cliName, prerunner, serverCompleter))
		cli.AddCommand(price.New(prerunner))
		cli.AddCommand(ps1.New(cliName, prerunner, &pps1.Prompt{}, logger))
		cli.AddCommand(schemaregistry.New(cliName, prerunner, nil, logger)) // Exposed for testing
		serviceAccountCmd := serviceaccount.New(prerunner)
		serverCompleter.AddCommand(serviceAccountCmd)
		cli.AddCommand(serviceAccountCmd.Command)
		cli.AddCommand(shell.NewShellCmd(cli, cfg, prerunner, shellCompleter, logger, analyticsClient, jwtValidator))
		cli.AddCommand(signup.New(prerunner, logger, ver.UserAgent))
		if os.Getenv("XX_CCLOUD_RBAC") != "" {
			cli.AddCommand(iam.New(cliName, prerunner))
		}
	} else if cliName == "confluent" {
		cli.AddCommand(auditlog.New(prerunner))
		cli.AddCommand(cluster.New(prerunner, cluster.NewScopedIdService(&http.Client{}, ver.UserAgent, logger)))
		cli.AddCommand(connect.New(prerunner))
		cli.AddCommand(iam.New(cliName, prerunner))
		// Never uses it under "confluent", so a nil ServerCompleter is fine.
		cli.AddCommand(kafka.New(isAPIKeyCredential(cfg), cliName, prerunner, logger.Named("kafka"), ver.ClientID, nil))
		cli.AddCommand(ksql.New(cliName, prerunner, nil))
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
	segmentClient, _ := segment.NewWithConfig(keys.SegmentKey, segment.Config{
		Logger: analytics.NewLogger(logger),
	})
	return analytics.NewAnalyticsClient(cliName, cfg, cliVersion, segmentClient, clockwork.NewRealClock())
}

func isAPIKeyCredential(cfg *v3.Config) bool {
	if cfg == nil {
		return false
	}
	currCtx := cfg.Context()
	return currCtx != nil && currCtx.Credential != nil && currCtx.Credential.CredentialType == v2.APIKey
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
