package internal

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	shell "github.com/brianstrauch/cobra-shell"
	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"

	"github.com/confluentinc/cli/v4/internal/ai"
	apikey "github.com/confluentinc/cli/v4/internal/api-key"
	"github.com/confluentinc/cli/v4/internal/asyncapi"
	auditlog "github.com/confluentinc/cli/v4/internal/audit-log"
	"github.com/confluentinc/cli/v4/internal/billing"
	"github.com/confluentinc/cli/v4/internal/byok"
	ccpm "github.com/confluentinc/cli/v4/internal/ccpm"
	cloudsignup "github.com/confluentinc/cli/v4/internal/cloud-signup"
	"github.com/confluentinc/cli/v4/internal/cluster"
	"github.com/confluentinc/cli/v4/internal/completion"
	"github.com/confluentinc/cli/v4/internal/configuration"
	"github.com/confluentinc/cli/v4/internal/connect"
	"github.com/confluentinc/cli/v4/internal/context"
	ccl "github.com/confluentinc/cli/v4/internal/custom-code-logging"
	"github.com/confluentinc/cli/v4/internal/environment"
	"github.com/confluentinc/cli/v4/internal/feedback"
	"github.com/confluentinc/cli/v4/internal/flink"
	"github.com/confluentinc/cli/v4/internal/iam"
	"github.com/confluentinc/cli/v4/internal/kafka"
	"github.com/confluentinc/cli/v4/internal/ksql"
	"github.com/confluentinc/cli/v4/internal/local"
	"github.com/confluentinc/cli/v4/internal/login"
	"github.com/confluentinc/cli/v4/internal/logout"
	"github.com/confluentinc/cli/v4/internal/network"
	"github.com/confluentinc/cli/v4/internal/organization"
	"github.com/confluentinc/cli/v4/internal/plugin"
	"github.com/confluentinc/cli/v4/internal/prompt"
	providerintegration "github.com/confluentinc/cli/v4/internal/provider-integration"
	schemaregistry "github.com/confluentinc/cli/v4/internal/schema-registry"
	"github.com/confluentinc/cli/v4/internal/secret"
	servicequota "github.com/confluentinc/cli/v4/internal/service-quota"
	streamshare "github.com/confluentinc/cli/v4/internal/stream-share"
	"github.com/confluentinc/cli/v4/internal/tableflow"
	unifiedstreammanager "github.com/confluentinc/cli/v4/internal/unified-stream-manager"
	"github.com/confluentinc/cli/v4/internal/update"
	"github.com/confluentinc/cli/v4/internal/version"
	pauth "github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
	"github.com/confluentinc/cli/v4/pkg/form"
	"github.com/confluentinc/cli/v4/pkg/help"
	"github.com/confluentinc/cli/v4/pkg/jwt"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
	ppanic "github.com/confluentinc/cli/v4/pkg/panic-recovery"
	pplugin "github.com/confluentinc/cli/v4/pkg/plugin"
	secrets "github.com/confluentinc/cli/v4/pkg/secret"
	"github.com/confluentinc/cli/v4/pkg/usage"
	pversion "github.com/confluentinc/cli/v4/pkg/version"
)

func NewConfluentCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     pversion.CLIName,
		Short:   fmt.Sprintf("%s.", pversion.FullCLIName),
		Long:    getLongDescription(cfg),
		Version: cfg.Version.Version,
	}

	cmd.Flags().Bool("version", false, fmt.Sprintf("Show version of the %s.", pversion.FullCLIName))
	cmd.PersistentFlags().BoolP("help", "h", false, "Show help for this command.")
	cmd.PersistentFlags().CountP("verbose", "v", "Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).")
	cmd.PersistentFlags().Bool("unsafe-trace", false, "Equivalent to -vvvv, but also log HTTP requests and responses which might contain plaintext secrets.")

	authTokenHandler := pauth.NewAuthTokenHandler()
	ccloudClientFactory := pauth.NewCCloudClientFactory(cfg.Version.UserAgent)
	jwtValidator := jwt.NewValidator()
	ccloudClient := getCloudClient(cfg, ccloudClientFactory)
	loginCredentialsManager := pauth.NewLoginCredentialsManager(form.NewPrompt(), ccloudClient)
	loginOrganizationManager := pauth.NewLoginOrganizationManagerImpl()
	mdsClientManager := &pauth.MDSClientManagerImpl{}
	featureflags.Init(cfg)

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		pcmd.LabelRequiredFlags(cmd)
		_ = help.WriteHelpTemplate(cmd)
	})

	prerunner := &pcmd.PreRun{
		Config:                  cfg,
		AuthTokenHandler:        authTokenHandler,
		CCloudClientFactory:     ccloudClientFactory,
		JWTValidator:            jwtValidator,
		LoginCredentialsManager: loginCredentialsManager,
		MDSClientManager:        mdsClientManager,
		Version:                 cfg.Version,
	}

	cmd.AddCommand(ai.New(prerunner))
	cmd.AddCommand(apikey.New(prerunner))
	cmd.AddCommand(asyncapi.New(prerunner))
	cmd.AddCommand(auditlog.New(prerunner))
	cmd.AddCommand(billing.New(prerunner))
	cmd.AddCommand(byok.New(prerunner))
	cmd.AddCommand(ccpm.New(cfg, prerunner))
	cmd.AddCommand(cluster.New(prerunner, cfg.Version.UserAgent))
	cmd.AddCommand(cloudsignup.New(prerunner))
	cmd.AddCommand(completion.New())
	cmd.AddCommand(configuration.New(cfg, prerunner))
	cmd.AddCommand(context.New(prerunner))
	cmd.AddCommand(connect.New(cfg, prerunner))
	cmd.AddCommand(ccl.New(cfg, prerunner))
	cmd.AddCommand(environment.New(prerunner))
	cmd.AddCommand(feedback.New(prerunner))
	cmd.AddCommand(flink.New(cfg, prerunner))
	cmd.AddCommand(iam.New(cfg, prerunner))
	cmd.AddCommand(kafka.New(cfg, prerunner))
	cmd.AddCommand(ksql.New(cfg, prerunner))
	cmd.AddCommand(local.New(cfg, prerunner))
	cmd.AddCommand(login.New(cfg, prerunner, ccloudClientFactory, mdsClientManager, loginCredentialsManager, loginOrganizationManager, authTokenHandler))
	cmd.AddCommand(logout.New(cfg, prerunner, authTokenHandler))
	cmd.AddCommand(network.New(prerunner, cfg))
	cmd.AddCommand(organization.New(prerunner))
	cmd.AddCommand(plugin.New(cfg, prerunner))
	cmd.AddCommand(prompt.New(cfg))
	cmd.AddCommand(providerintegration.New(prerunner))
	cmd.AddCommand(servicequota.New(prerunner))
	cmd.AddCommand(schemaregistry.New(cfg, prerunner))
	cmd.AddCommand(secret.New(prerunner, secrets.NewPasswordProtectionPlugin()))
	cmd.AddCommand(shell.New(cmd, func() *cobra.Command { return NewConfluentCommand(cfg) }))
	cmd.AddCommand(streamshare.New(prerunner))
	cmd.AddCommand(tableflow.New(prerunner))
	cmd.AddCommand(unifiedstreammanager.New(prerunner, cfg))
	cmd.AddCommand(update.New(cfg, prerunner))
	cmd.AddCommand(version.New(prerunner, cfg.Version))

	_ = cfg.ParseFlagsIntoConfig(cmd)

	changeDefaults(cmd, cfg)
	deprecateCommandsAndFlags(cmd, cfg)
	featureflags.Manager.SetCommandAndFlags(cmd, os.Args[1:])
	disableCommandAndFlagHelpText(cmd, cfg)
	return cmd
}

func Execute(cmd *cobra.Command, args []string, cfg *config.Config) error {
	defer func() {
		if r := recover(); r != nil {
			if !cfg.IsCloudLogin() || cfg.HasGovHostname() || !cfg.Version.IsReleased() {
				log.CliLogger.Trace(string(debug.Stack()))
			}
			u := ppanic.CollectPanic(cmd, args, cfg)
			if err := reportUsage(cmd, cfg, u); err != nil {
				output.ErrPrint(cfg.EnableColor, errors.DisplaySuggestionsMessage(err))
			}
			cobra.CheckErr(r)
		}
	}()
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
	output.ErrPrint(cfg.EnableColor, errors.DisplaySuggestionsMessage(err))

	u.Error = cliv1.PtrBool(err != nil)
	if err := reportUsage(cmd, cfg, u); err != nil {
		return err
	}

	return err
}

func reportUsage(cmd *cobra.Command, cfg *config.Config, u *usage.Usage) error {
	if cfg.IsCloudLogin() && !cfg.HasGovHostname() && u.Command != nil && *u.Command != "" {
		unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
		if err != nil {
			return err
		}
		u.Report(ccloudv2.NewClient(cfg, unsafeTrace))
	}
	return nil
}

func getLongDescription(cfg *config.Config) string {
	switch {
	case cfg.IsCloudLogin():
		return "Manage your Confluent Cloud."
	case cfg.IsOnPremLogin():
		return "Manage your Confluent Platform."
	default:
		return "Manage your Confluent Cloud or Confluent Platform. Log in to see all available commands."
	}
}

func changeDefaults(cmd *cobra.Command, cfg *config.Config) {
	hideAndErrIfMissingRunRequirement(cmd, cfg)
	catchErrors(cmd)

	cmd.Flags().SortFlags = false

	for _, subcommand := range cmd.Commands() {
		changeDefaults(subcommand, cfg)
	}
}

// hideAndErrIfMissingRunRequirement hides commands that don't meet a requirement and errs if a user attempts to use it;
// for example, an on-prem command shouldn't be used by a cloud user.
func hideAndErrIfMissingRunRequirement(cmd *cobra.Command, cfg *config.Config) {
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

func getCloudClient(cfg *config.Config, ccloudClientFactory pauth.CCloudClientFactory) *ccloudv1.Client {
	if cfg.IsCloudLogin() {
		return ccloudClientFactory.AnonHTTPClientFactory(pauth.CCloudURL)
	}
	return nil
}

func deprecateCommandsAndFlags(cmd *cobra.Command, cfg *config.Config) {
	deprecatedCmds := featureflags.Manager.JsonVariation(featureflags.DeprecationNotices, cfg.Context(), config.CliLaunchDarklyClient, true, []any{})
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

func disableCommandAndFlagHelpText(cmd *cobra.Command, cfg *config.Config) {
	disableResp := featureflags.GetLDDisableMap(cfg.Context())
	disabledCmdsAndFlags, ok := disableResp["patterns"].([]any)
	if ok && len(disabledCmdsAndFlags) > 0 {
		for _, pattern := range disabledCmdsAndFlags {
			if command, flags, err := cmd.Find(strings.Split(pattern.(string), " ")); err == nil {
				featureflags.DisableHelpText(command, flags)
			}
		}
	}
}
