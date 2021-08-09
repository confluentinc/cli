package cmd

import (
	"fmt"
	"os"

	"github.com/jonboulle/clockwork"
	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/admin"
	"github.com/confluentinc/cli/internal/cmd/api-key"
	"github.com/confluentinc/cli/internal/cmd/audit-log"
	"github.com/confluentinc/cli/internal/cmd/cluster"
	"github.com/confluentinc/cli/internal/cmd/completion"
	"github.com/confluentinc/cli/internal/cmd/config"
	"github.com/confluentinc/cli/internal/cmd/connect"
	"github.com/confluentinc/cli/internal/cmd/connector"
	"github.com/confluentinc/cli/internal/cmd/connector-catalog"
	"github.com/confluentinc/cli/internal/cmd/environment"
	"github.com/confluentinc/cli/internal/cmd/iam"
	initcontext "github.com/confluentinc/cli/internal/cmd/init"
	"github.com/confluentinc/cli/internal/cmd/kafka"
	"github.com/confluentinc/cli/internal/cmd/ksql"
	"github.com/confluentinc/cli/internal/cmd/local"
	"github.com/confluentinc/cli/internal/cmd/login"
	"github.com/confluentinc/cli/internal/cmd/logout"
	"github.com/confluentinc/cli/internal/cmd/price"
	"github.com/confluentinc/cli/internal/cmd/prompt"
	"github.com/confluentinc/cli/internal/cmd/schema-registry"
	"github.com/confluentinc/cli/internal/cmd/secret"
	"github.com/confluentinc/cli/internal/cmd/service-account"
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
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/help"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/metric"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/ps1"
	secrets "github.com/confluentinc/cli/internal/pkg/secret"
	"github.com/confluentinc/cli/internal/pkg/shell/completer"
	keys "github.com/confluentinc/cli/internal/pkg/third-party-keys"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	"github.com/confluentinc/cli/mock"
)

type command struct {
	*cobra.Command
	// @VisibleForTesting
	Analytics analytics.Client
	logger    *log.Logger
}

func NewConfluentCommand(cfg *v3.Config, isTest bool, ver *pversion.Version) *command {
	cli := &cobra.Command{
		Use:               pversion.CLIName,
		Short:             fmt.Sprintf("%s.", pversion.FullCLIName),
		Long:              getLongDescription(cfg),
		Version:           ver.Version,
		DisableAutoGenTag: true,
	}

	cli.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		pcmd.LabelRequiredFlags(cmd)
		_ = help.WriteHelpTemplate(cmd)
	})

	cli.PersistentFlags().BoolP("help", "h", false, "Show help for this command.")
	cli.PersistentFlags().CountP("verbose", "v", "Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).")
	cli.Flags().Bool("version", false, fmt.Sprintf("Show version of the %s.", pversion.FullCLIName))

	logger := log.New()

	disableUpdateCheck := cfg.DisableUpdates || cfg.DisableUpdateCheck
	updateClient := update.NewClient(pversion.CLIName, disableUpdateCheck, logger)

	analyticsClient := getAnalyticsClient(isTest, pversion.CLIName, cfg, ver.Version, logger)
	authTokenHandler := pauth.NewAuthTokenHandler(logger)
	ccloudClientFactory := pauth.NewCCloudClientFactory(ver.UserAgent, logger)
	flagResolver := &pcmd.FlagResolverImpl{Prompt: form.NewPrompt(os.Stdin), Out: os.Stdout}
	jwtValidator := pcmd.NewJWTValidator(logger)
	netrcHandler := netrc.NewNetrcHandler(netrc.GetNetrcFilePath(isTest))
	loginCredentialsManager := pauth.NewLoginCredentialsManager(netrcHandler, form.NewPrompt(os.Stdin), logger)
	mdsClientManager := &pauth.MDSClientManagerImpl{}

	prerunner := &pcmd.PreRun{
		Analytics:               analyticsClient,
		AuthTokenHandler:        authTokenHandler,
		CCloudClientFactory:     ccloudClientFactory,
		Config:                  cfg,
		FlagResolver:            flagResolver,
		IsTest:                  isTest,
		JWTValidator:            jwtValidator,
		Logger:                  logger,
		LoginCredentialsManager: loginCredentialsManager,
		MDSClientManager:        mdsClientManager,
		UpdateClient:            updateClient,
		Version:                 ver,
	}

	command := &command{Command: cli, Analytics: analyticsClient, logger: logger}

	var serverCompleter completer.ServerSideCompleter
	shellCompleter := completer.NewShellCompleter(cli)
	if cfg.IsCloud() {
		serverCompleter = shellCompleter.ServerSideCompleter
	}

	isAPIKeyLogin := isAPIKeyCredential(cfg)

	// No-login commands
	cli.AddCommand(completion.New(cli))
	cli.AddCommand(config.New(cfg.IsCloud(), prerunner, analyticsClient))
	cli.AddCommand(kafka.New(cfg, isAPIKeyLogin, prerunner, logger.Named("kafka"), ver.ClientID, serverCompleter, analyticsClient))
	cli.AddCommand(local.New(prerunner))
	cli.AddCommand(login.New(prerunner, logger, ccloudClientFactory, mdsClientManager, analyticsClient, netrcHandler, loginCredentialsManager, authTokenHandler, isTest).Command)
	cli.AddCommand(logout.New(cfg, prerunner, analyticsClient, netrcHandler).Command)
	cli.AddCommand(secret.New(flagResolver, secrets.NewPasswordProtectionPlugin(logger)))
	if !cfg.DisableUpdates {
		cli.AddCommand(update.New(logger, ver, updateClient, analyticsClient))
	}
	cli.AddCommand(version.New(prerunner, ver))

	if cfg.IsCloud() {
		cli.AddCommand(admin.New(prerunner, isTest))
		cli.AddCommand(auditlog.New(cfg, prerunner))
		cli.AddCommand(initcontext.New(prerunner, flagResolver, analyticsClient))

		// If a user logs in with an API key, don't allow the remaining commands.
		if isAPIKeyLogin {
			return command
		}

		apiKeyCmd := apikey.New(prerunner, nil, flagResolver, analyticsClient)
		connectorCmd := connector.New(prerunner, analyticsClient)
		connectorCatalogCmd := connectorcatalog.New(prerunner)
		environmentCmd := environment.New(prerunner, analyticsClient)
		serviceAccountCmd := serviceaccount.New(prerunner, analyticsClient)

		serverCompleter.AddCommand(apiKeyCmd)
		serverCompleter.AddCommand(connectorCmd)
		serverCompleter.AddCommand(connectorCatalogCmd)
		serverCompleter.AddCommand(environmentCmd)
		serverCompleter.AddCommand(serviceAccountCmd)

		cli.AddCommand(apiKeyCmd.Command)
		cli.AddCommand(connectorCatalogCmd.Command)
		cli.AddCommand(connectorCmd.Command)
		cli.AddCommand(environmentCmd.Command)
		cli.AddCommand(iam.New(cfg, prerunner))
		cli.AddCommand(ksql.New(cfg, prerunner, serverCompleter, analyticsClient))
		cli.AddCommand(price.New(prerunner))
		cli.AddCommand(prompt.New(cfg, prerunner, &ps1.Prompt{}, logger))
		cli.AddCommand(schemaregistry.New(cfg, prerunner, nil, logger, analyticsClient))
		cli.AddCommand(serviceAccountCmd.Command)
		cli.AddCommand(shell.NewShellCmd(cli, prerunner, cfg, shellCompleter, jwtValidator))
		cli.AddCommand(signup.New(prerunner, logger, ver.UserAgent, ccloudClientFactory).Command)
	}

	if cfg.IsOnPrem() {
		cli.AddCommand(auditlog.New(cfg, prerunner))
		cli.AddCommand(cluster.New(prerunner, cluster.NewScopedIdService(ver.UserAgent, logger)))
		cli.AddCommand(connect.New(prerunner))
		cli.AddCommand(iam.New(cfg, prerunner))
		cli.AddCommand(ksql.New(cfg, prerunner, serverCompleter, analyticsClient))
		cli.AddCommand(schemaregistry.New(cfg, prerunner, nil, logger, analyticsClient))
	}

	return command
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
	ctx := cfg.Context()
	return ctx != nil && ctx.Credential != nil && ctx.Credential.CredentialType == v2.APIKey
}

func (c *command) Execute(args []string) error {
	c.Analytics.SetStartTime()
	c.Command.SetArgs(args)
	err := c.Command.Execute()
	errors.DisplaySuggestionsMessage(err, os.Stderr)
	c.sendAndFlushAnalytics(args, err)
	return err
}

func (c *command) sendAndFlushAnalytics(args []string, err error) {
	if err := c.Analytics.SendCommandAnalytics(c.Command, args, err); err != nil {
		c.logger.Debugf("segment analytics sending event failed: %s\n", err.Error())
	}

	if err := c.Analytics.Close(); err != nil {
		c.logger.Debug(err)
	}
}

func LoadConfig() (*v3.Config, error) {
	cfg := v3.New(&pconfig.Params{
		Logger:     log.New(),
		MetricSink: metric.NewSink(),
	})

	return load.LoadAndMigrate(cfg)
}

func getLongDescription(cfg *v3.Config) string {
	if cfg.IsCloud() {
		return "Manage your Confluent Cloud."
	}

	if cfg.IsOnPrem() {
		return "Manage your Confluent Platform."
	}

	return "Manage your Confluent Cloud or Confluent Platform. Log in to see all available commands."
}
