package cmd

import (
	"fmt"
	"os"

	shell "github.com/brianstrauch/cobra-shell"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/admin"
	apikey "github.com/confluentinc/cli/internal/cmd/api-key"
	auditlog "github.com/confluentinc/cli/internal/cmd/audit-log"
	cloudsignup "github.com/confluentinc/cli/internal/cmd/cloud-signup"
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
	schemaregistry "github.com/confluentinc/cli/internal/cmd/schema-registry"
	"github.com/confluentinc/cli/internal/cmd/secret"
	"github.com/confluentinc/cli/internal/cmd/update"
	"github.com/confluentinc/cli/internal/cmd/version"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	pconfig "github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/config/load"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/help"
	"github.com/confluentinc/cli/internal/pkg/metric"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	secrets "github.com/confluentinc/cli/internal/pkg/secret"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

type command struct {
	*cobra.Command
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

	disableUpdateCheck := cfg.DisableUpdates || cfg.DisableUpdateCheck
	updateClient := update.NewClient(pversion.CLIName, disableUpdateCheck)

	authTokenHandler := pauth.NewAuthTokenHandler()
	ccloudClientFactory := pauth.NewCCloudClientFactory(ver.UserAgent)
	flagResolver := &pcmd.FlagResolverImpl{Prompt: form.NewPrompt(os.Stdin), Out: os.Stdout}
	jwtValidator := pcmd.NewJWTValidator()
	netrcHandler := netrc.NewNetrcHandler(netrc.GetNetrcFilePath(isTest))
	loginCredentialsManager := pauth.NewLoginCredentialsManager(netrcHandler, form.NewPrompt(os.Stdin), getCloudClient(cfg, ccloudClientFactory))
	mdsClientManager := &pauth.MDSClientManagerImpl{}

	prerunner := &pcmd.PreRun{
		AuthTokenHandler:        authTokenHandler,
		CCloudClientFactory:     ccloudClientFactory,
		Config:                  cfg,
		FlagResolver:            flagResolver,
		IsTest:                  isTest,
		JWTValidator:            jwtValidator,
		LoginCredentialsManager: loginCredentialsManager,
		MDSClientManager:        mdsClientManager,
		UpdateClient:            updateClient,
		Version:                 ver,
	}

	cmd.AddCommand(admin.New(prerunner, isTest))
	cmd.AddCommand(apikey.New(prerunner, nil, flagResolver))
	cmd.AddCommand(auditlog.New(prerunner))
	cmd.AddCommand(cluster.New(prerunner, ver.UserAgent))
	cmd.AddCommand(cloudsignup.New(prerunner, ver.UserAgent, ccloudClientFactory).Command)
	cmd.AddCommand(completion.New())
	cmd.AddCommand(context.New(prerunner, flagResolver))
	cmd.AddCommand(connect.New(prerunner))
	cmd.AddCommand(environment.New(prerunner))
	cmd.AddCommand(iam.New(cfg, prerunner))
	cmd.AddCommand(kafka.New(cfg, prerunner, ver.ClientID))
	cmd.AddCommand(ksql.New(cfg, prerunner))
	cmd.AddCommand(local.New(prerunner))
	cmd.AddCommand(login.New(cfg, prerunner, ccloudClientFactory, mdsClientManager, netrcHandler, loginCredentialsManager, authTokenHandler).Command)
	cmd.AddCommand(logout.New(cfg, prerunner, netrcHandler).Command)
	cmd.AddCommand(price.New(prerunner))
	cmd.AddCommand(prompt.New(cfg))
	cmd.AddCommand(schemaregistry.New(cfg, prerunner, nil))
	cmd.AddCommand(secret.New(prerunner, flagResolver, secrets.NewPasswordProtectionPlugin()))
	cmd.AddCommand(shell.New(cmd))
	cmd.AddCommand(update.New(prerunner, ver, updateClient))
	cmd.AddCommand(version.New(prerunner, ver))

	hideAndErrIfMissingRunRequirement(cmd, cfg)
	disableFlagSorting(cmd)

	return &command{Command: cmd}
}

func (c *command) Execute(args []string) error {
	c.Command.SetArgs(args)
	err := c.Command.Execute()
	errors.DisplaySuggestionsMessage(err, os.Stderr)
	return err
}

func LoadConfig() (*v1.Config, error) {
	cfg := v1.New(&pconfig.Params{
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
