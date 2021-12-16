package cmd

import (
	"fmt"
	"os"

	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/jonboulle/clockwork"
	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/admin"
	"github.com/confluentinc/cli/internal/cmd/api-key"
	"github.com/confluentinc/cli/internal/cmd/audit-log"
	"github.com/confluentinc/cli/internal/cmd/cloud-signup"
	"github.com/confluentinc/cli/internal/cmd/cluster"
	"github.com/confluentinc/cli/internal/cmd/completion"
	"github.com/confluentinc/cli/internal/cmd/connect"
	"github.com/confluentinc/cli/internal/cmd/context"
	"github.com/confluentinc/cli/internal/cmd/environment"
	"github.com/confluentinc/cli/internal/cmd/iam"
	"github.com/confluentinc/cli/internal/cmd/kafka"
	"github.com/confluentinc/cli/internal/cmd/ksql"
	"github.com/confluentinc/cli/internal/cmd/local"
	"github.com/confluentinc/cli/internal/cmd/login"
	"github.com/confluentinc/cli/internal/cmd/logout"
	"github.com/confluentinc/cli/internal/cmd/price"
	"github.com/confluentinc/cli/internal/cmd/prompt"
	"github.com/confluentinc/cli/internal/cmd/schema-registry"
	"github.com/confluentinc/cli/internal/cmd/secret"
	"github.com/confluentinc/cli/internal/cmd/shell"
	"github.com/confluentinc/cli/internal/cmd/update"
	"github.com/confluentinc/cli/internal/cmd/version"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	pconfig "github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/config/load"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/help"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/metric"
	"github.com/confluentinc/cli/internal/pkg/netrc"
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

func NewConfluentCommand(cfg *v1.Config, isTest bool, ver *pversion.Version) *command {
	cmd := &cobra.Command{
		Use:     pversion.CLIName,
		Short:   fmt.Sprintf("%s.", pversion.FullCLIName),
		Long:    getLongDescription(cfg),
		Version: ver.Version,
	}

	cmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		pcmd.LabelRequiredFlags(cmd)
		_ = help.WriteHelpTemplate(cmd)
	})

	cmd.Flags().Bool("version", false, fmt.Sprintf("Show version of the %s.", pversion.FullCLIName))
	cmd.PersistentFlags().BoolP("help", "h", false, "Show help for this command.")
	cmd.PersistentFlags().CountP("verbose", "v", "Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).")

	logger := log.New()

	disableUpdateCheck := cfg.DisableUpdates || cfg.DisableUpdateCheck
	updateClient := update.NewClient(pversion.CLIName, disableUpdateCheck, logger)

	analyticsClient := getAnalyticsClient(isTest, cfg, ver.Version, logger)
	authTokenHandler := pauth.NewAuthTokenHandler(logger)
	ccloudClientFactory := pauth.NewCCloudClientFactory(ver.UserAgent, logger)
	flagResolver := &pcmd.FlagResolverImpl{Prompt: form.NewPrompt(os.Stdin), Out: os.Stdout}
	jwtValidator := pcmd.NewJWTValidator(logger)
	netrcHandler := netrc.NewNetrcHandler(netrc.GetNetrcFilePath(isTest))
	loginCredentialsManager := pauth.NewLoginCredentialsManager(netrcHandler, form.NewPrompt(os.Stdin), logger, getCloudClient(cfg, ccloudClientFactory))
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

	var serverCompleter completer.ServerSideCompleter
	shellCompleter := completer.NewShellCompleter(cmd)
	if cfg.IsCloudLogin() {
		serverCompleter = shellCompleter.ServerSideCompleter
	}

	apiKeyCmd := apikey.New(prerunner, nil, flagResolver, analyticsClient)
	connectCmd := connect.New(prerunner, analyticsClient)
	environmentCmd := environment.New(prerunner, analyticsClient)

	cmd.AddCommand(admin.New(prerunner, isTest))
	cmd.AddCommand(apiKeyCmd.Command)
	cmd.AddCommand(auditlog.New(prerunner))
	cmd.AddCommand(cluster.New(prerunner, ver.UserAgent, logger))
	cmd.AddCommand(cloudsignup.New(prerunner, logger, ver.UserAgent, ccloudClientFactory).Command)
	cmd.AddCommand(completion.New())
	cmd.AddCommand(context.New(prerunner, flagResolver))
	cmd.AddCommand(connectCmd.Command)
	cmd.AddCommand(environmentCmd.Command)
	cmd.AddCommand(iam.New(cfg, prerunner, serverCompleter))
	cmd.AddCommand(kafka.New(cfg, prerunner, logger.Named("kafka"), ver.ClientID, serverCompleter, analyticsClient))
	cmd.AddCommand(ksql.New(cfg, prerunner, serverCompleter, analyticsClient))
	cmd.AddCommand(local.New(prerunner))
	cmd.AddCommand(login.New(prerunner, logger, ccloudClientFactory, mdsClientManager, analyticsClient, netrcHandler, loginCredentialsManager, authTokenHandler, isTest).Command)
	cmd.AddCommand(logout.New(cfg, prerunner, analyticsClient, netrcHandler).Command)
	cmd.AddCommand(price.New(prerunner))
	cmd.AddCommand(prompt.New(cfg))
	cmd.AddCommand(schemaregistry.New(cfg, prerunner, nil, logger, analyticsClient))
	cmd.AddCommand(secret.New(prerunner, flagResolver, secrets.NewPasswordProtectionPlugin(logger)))
	cmd.AddCommand(shell.NewShellCmd(cmd, prerunner, cfg, shellCompleter, jwtValidator))
	cmd.AddCommand(update.New(prerunner, logger, ver, updateClient, analyticsClient))
	cmd.AddCommand(version.New(prerunner, ver))

	if cfg.IsCloudLogin() {
		serverCompleter.AddCommand(apiKeyCmd)
		serverCompleter.AddCommand(connectCmd)
		serverCompleter.AddCommand(environmentCmd)
	}

	hideAndErrIfMissingRunRequirement(cmd, cfg)
	disableFlagSorting(cmd)

	return &command{Command: cmd, Analytics: analyticsClient, logger: logger}
}

func getAnalyticsClient(isTest bool, cfg *v1.Config, cliVersion string, logger *log.Logger) analytics.Client {
	if cfg.IsOnPremLogin() || isTest {
		return mock.NewDummyAnalyticsMock()
	}
	segmentClient, _ := segment.NewWithConfig(keys.SegmentKey, segment.Config{
		Logger: analytics.NewLogger(logger),
	})
	return analytics.NewAnalyticsClient(cfg, cliVersion, segmentClient, clockwork.NewRealClock())
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

func LoadConfig() (*v1.Config, error) {
	cfg := v1.New(&pconfig.Params{
		Logger:     log.New(),
		MetricSink: metric.NewSink(),
	})

	return load.LoadAndMigrate(cfg)
}

func getLongDescription(cfg *v1.Config) string {
	switch {
	case cfg.IsCloudLogin():
		return "Manage your Confluent Cloud."
	case cfg.IsOnPremLogin():
		return "Manage your Confluent Platform."
	default:
		return "Manage your Confluent Cloud or Confluent Platform. Log in to see all available commands."
	}
}

// hideAndErrIfMissingRunRequirement hides commands that don't meet a requirement and errs if a user attempts to use it;
// for example, an on-prem command shouldn't be used by a cloud user.
func hideAndErrIfMissingRunRequirement(cmd *cobra.Command, cfg *v1.Config) {
	if err := pcmd.ErrIfMissingRunRequirement(cmd, cfg); err != nil {
		cmd.Hidden = true

		// Show err for internal commands. Leaf commands will err in the PreRun function.
		if cmd.HasSubCommands() {
			cmd.RunE = func(_ *cobra.Command, _ []string) error { return err }
			cmd.SilenceUsage = true
		}
	}

	for _, subcommand := range cmd.Commands() {
		hideAndErrIfMissingRunRequirement(subcommand, cfg)
	}
}

// disableFlagSorting recursively disables the default option to sort flags, for all commands.
func disableFlagSorting(cmd *cobra.Command) {
	cmd.Flags().SortFlags = false

	for _, subcommand := range cmd.Commands() {
		disableFlagSorting(subcommand)
	}
}

func getCloudClient(cfg *v1.Config, ccloudClientFactory pauth.CCloudClientFactory) *ccloud.Client {
	if cfg.IsCloudLogin() {
		return ccloudClientFactory.AnonHTTPClientFactory(pauth.CCloudURL)
	}
	return nil
}
