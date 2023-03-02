package cmd

import (
	"fmt"
	"os"
	"strings"

	shell "github.com/brianstrauch/cobra-shell"
	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/admin"
	apikey "github.com/confluentinc/cli/internal/cmd/api-key"
	"github.com/confluentinc/cli/internal/cmd/asyncapi"
	auditlog "github.com/confluentinc/cli/internal/cmd/audit-log"
	byok "github.com/confluentinc/cli/internal/cmd/byok"
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
	"github.com/confluentinc/cli/internal/cmd/pipeline"
	"github.com/confluentinc/cli/internal/cmd/plugin"
	"github.com/confluentinc/cli/internal/cmd/price"
	"github.com/confluentinc/cli/internal/cmd/prompt"
	schemaregistry "github.com/confluentinc/cli/internal/cmd/schema-registry"
	"github.com/confluentinc/cli/internal/cmd/secret"
	servicequota "github.com/confluentinc/cli/internal/cmd/service-quota"
	streamshare "github.com/confluentinc/cli/internal/cmd/stream-share"
	"github.com/confluentinc/cli/internal/cmd/update"
	"github.com/confluentinc/cli/internal/cmd/version"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/featureflags"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/help"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/output"
	pplugin "github.com/confluentinc/cli/internal/pkg/plugin"
	secrets "github.com/confluentinc/cli/internal/pkg/secret"
	"github.com/confluentinc/cli/internal/pkg/usage"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

func NewConfluentCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     pversion.CLIName,
		Short:   fmt.Sprintf("%s.", pversion.FullCLIName),
		Long:    getLongDescription(cfg),
		Version: cfg.Version.Version,
	}

	cmd.Flags().Bool("version", false, fmt.Sprintf("Show version of the %s.", pversion.FullCLIName))
	cmd.PersistentFlags().BoolP("help", "h", false, "Show help for this command.")
	cmd.PersistentFlags().CountP("verbose", "v", "Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).")
	cmd.PersistentFlags().Bool("unsafe-trace", false, "Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.")

	disableUpdateCheck := cfg.DisableUpdates || cfg.DisableUpdateCheck
	updateClient := update.NewClient(pversion.CLIName, disableUpdateCheck)
	authTokenHandler := pauth.NewAuthTokenHandler()
	ccloudClientFactory := pauth.NewCCloudClientFactory(cfg.Version.UserAgent)
	flagResolver := &pcmd.FlagResolverImpl{Prompt: form.NewPrompt(os.Stdin), Out: os.Stdout}
	jwtValidator := pcmd.NewJWTValidator()
	netrcHandler := netrc.NewNetrcHandler(netrc.GetNetrcFilePath(cfg.IsTest))
	ccloudClient := getCloudClient(cfg, ccloudClientFactory)
	loginCredentialsManager := pauth.NewLoginCredentialsManager(netrcHandler, form.NewPrompt(os.Stdin), ccloudClient)
	loginOrganizationManager := pauth.NewLoginOrganizationManagerImpl()
	mdsClientManager := &pauth.MDSClientManagerImpl{}
	featureflags.Init(cfg.Version, cfg.IsTest, cfg.DisableFeatureFlags)

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		pcmd.LabelRequiredFlags(cmd)
		_ = help.WriteHelpTemplate(cmd)
	})

	prerunner := &pcmd.PreRun{
		Config:                  cfg,
		AuthTokenHandler:        authTokenHandler,
		CCloudClientFactory:     ccloudClientFactory,
		FlagResolver:            flagResolver,
		JWTValidator:            jwtValidator,
		LoginCredentialsManager: loginCredentialsManager,
		MDSClientManager:        mdsClientManager,
		UpdateClient:            updateClient,
		Version:                 cfg.Version,
	}

	cmd.AddCommand(admin.New(prerunner, cfg.IsTest))
	cmd.AddCommand(apikey.New(prerunner, nil, flagResolver))
	cmd.AddCommand(asyncapi.New(prerunner))
	cmd.AddCommand(auditlog.New(prerunner))
	cmd.AddCommand(byok.New(prerunner))
	cmd.AddCommand(cluster.New(prerunner, cfg.Version.UserAgent))
	cmd.AddCommand(cloudsignup.New())
	cmd.AddCommand(completion.New())
	cmd.AddCommand(context.New(prerunner, flagResolver))
	cmd.AddCommand(connect.New(cfg, prerunner))
	cmd.AddCommand(environment.New(prerunner))
	cmd.AddCommand(iam.New(cfg, prerunner))
	cmd.AddCommand(kafka.New(cfg, prerunner, cfg.Version.ClientID))
	cmd.AddCommand(ksql.New(cfg, prerunner))
	cmd.AddCommand(local.New(prerunner))
	cmd.AddCommand(login.New(cfg, prerunner, ccloudClientFactory, mdsClientManager, netrcHandler, loginCredentialsManager, loginOrganizationManager, authTokenHandler))
	cmd.AddCommand(logout.New(cfg, prerunner, netrcHandler))
	cmd.AddCommand(pipeline.New(cfg, prerunner))
	cmd.AddCommand(plugin.New(cfg, prerunner))
	cmd.AddCommand(price.New(prerunner))
	cmd.AddCommand(prompt.New(cfg))
	cmd.AddCommand(servicequota.New(prerunner))
	cmd.AddCommand(schemaregistry.New(cfg, prerunner, nil))
	cmd.AddCommand(secret.New(prerunner, flagResolver, secrets.NewPasswordProtectionPlugin()))
	cmd.AddCommand(shell.New(cmd, func() *cobra.Command { return NewConfluentCommand(cfg) }))
	cmd.AddCommand(update.New(prerunner, cfg.Version, updateClient))
	cmd.AddCommand(version.New(prerunner, cfg.Version))

	dc := dynamicconfig.New(cfg, nil, nil)
	_ = dc.ParseFlagsIntoConfig(cmd)
	if cfg.IsTest || featureflags.Manager.BoolVariation("cli.cdx", dc.Context(), v1.CliLaunchDarklyClient, true, false) {
		cmd.AddCommand(streamshare.New(cfg, prerunner))
	}

	changeDefaults(cmd, cfg)
	deprecateCommandsAndFlags(cmd, cfg)
	return cmd
}

func Execute(cmd *cobra.Command, args []string, cfg *v1.Config) error {
	if !cfg.DisablePlugins {
		if plugin := pplugin.FindPlugin(cmd, args, cfg); plugin != nil {
			return pplugin.ExecPlugin(plugin)
		}
	}
	// Usage collection is a wrapper around Execute() instead of a post-run function so we can collect the error status.
	u := usage.New(cfg.Version.Version)

	if !cfg.IsTest && cfg.Version.IsReleased() {
		cmd.PersistentPostRun = u.Collect
	}

	err := cmd.Execute()
	output.ErrPrint(errors.DisplaySuggestionsMessage(err))
	u.Error = cliv1.PtrBool(err != nil)

	if cfg.IsCloudLogin() && u.Command != nil && *(u.Command) != "" {
		unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
		if err != nil {
			return err
		}
		u.Report(cfg.GetCloudClientV2(unsafeTrace))
	}

	return err
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

func changeDefaults(cmd *cobra.Command, cfg *v1.Config) {
	hideAndErrIfMissingRunRequirement(cmd, cfg)
	catchErrors(cmd)

	cmd.Flags().SortFlags = false

	for _, subcommand := range cmd.Commands() {
		changeDefaults(subcommand, cfg)
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
		}
	}
}

// catchErrors catches (and modifies) errors from any of the built-in error-producing functions.
func catchErrors(cmd *cobra.Command) {
	if cmd.PersistentPreRunE != nil {
		cmd.PersistentPreRunE = pcmd.CatchErrors(cmd.PersistentPreRunE)
	}
	if cmd.PreRunE != nil {
		cmd.PreRunE = pcmd.CatchErrors(cmd.PreRunE)
	}
	if cmd.RunE != nil {
		cmd.RunE = pcmd.CatchErrors(cmd.RunE)
	}
}

func getCloudClient(cfg *v1.Config, ccloudClientFactory pauth.CCloudClientFactory) *ccloudv1.Client {
	if cfg.IsCloudLogin() {
		return ccloudClientFactory.AnonHTTPClientFactory(pauth.CCloudURL)
	}
	return nil
}

func deprecateCommandsAndFlags(cmd *cobra.Command, cfg *v1.Config) {
	ctx := dynamicconfig.NewDynamicContext(cfg.Context(), nil, nil)
	deprecatedCmds := featureflags.Manager.JsonVariation(featureflags.DeprecationNotices, ctx, v1.CliLaunchDarklyClient, true, []any{})
	cmdToFlagsAndMsg := featureflags.GetAnnouncementsOrDeprecation(deprecatedCmds)
	for name, flagsAndMsg := range cmdToFlagsAndMsg {
		if cmd, _, err := cmd.Find(strings.Split(name, " ")); err == nil {
			if flagsAndMsg.Flags == nil {
				featureflags.DeprecateCommandTree(cmd)
			} else {
				featureflags.DeprecateFlags(cmd, flagsAndMsg.Flags)
			}
		}
	}
}
