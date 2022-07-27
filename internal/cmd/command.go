package cmd

import (
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"os"
	"path/filepath"
	"strings"

	shell "github.com/brianstrauch/cobra-shell"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"
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
	"github.com/confluentinc/cli/internal/cmd/plugin"
	"github.com/confluentinc/cli/internal/cmd/price"
	"github.com/confluentinc/cli/internal/cmd/prompt"
	schemaregistry "github.com/confluentinc/cli/internal/cmd/schema-registry"
	"github.com/confluentinc/cli/internal/cmd/secret"
	servicequota "github.com/confluentinc/cli/internal/cmd/service-quota"
	"github.com/confluentinc/cli/internal/cmd/update"
	"github.com/confluentinc/cli/internal/cmd/version"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/featureflags"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/help"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	pplugin "github.com/confluentinc/cli/internal/pkg/plugin"
	secrets "github.com/confluentinc/cli/internal/pkg/secret"
	"github.com/confluentinc/cli/internal/pkg/usage"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

func NewConfluentCommand(cfg *v1.Config, ver *pversion.Version, isTest bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:     pversion.CLIName,
		Short:   fmt.Sprintf("%s.", pversion.FullCLIName),
		Long:    getLongDescription(cfg),
		Version: ver.Version,
	}

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
	loginOrganizationManager := pauth.NewLoginOrganizationManagerImpl()
	mdsClientManager := &pauth.MDSClientManagerImpl{}
	featureflags.Init(ver, isTest)

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		pcmd.LabelRequiredFlags(cmd)
		_ = help.WriteHelpTemplate(cmd)
	})

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
	cmd.AddCommand(cloudsignup.New(prerunner, ver.UserAgent, ccloudClientFactory, isTest))
	cmd.AddCommand(completion.New())
	cmd.AddCommand(context.New(prerunner, flagResolver))
	cmd.AddCommand(connect.New(prerunner))
	cmd.AddCommand(environment.New(prerunner))
	cmd.AddCommand(iam.New(cfg, prerunner))
	cmd.AddCommand(kafka.New(cfg, prerunner, ver.ClientID))
	cmd.AddCommand(ksql.New(cfg, prerunner))
	cmd.AddCommand(local.New(prerunner))
	cmd.AddCommand(login.New(cfg, prerunner, ccloudClientFactory, mdsClientManager, netrcHandler, loginCredentialsManager, loginOrganizationManager, authTokenHandler, isTest))
	cmd.AddCommand(logout.New(cfg, prerunner, netrcHandler))
	cmd.AddCommand(plugin.New(prerunner))
	cmd.AddCommand(price.New(prerunner))
	cmd.AddCommand(prompt.New(cfg))
	cmd.AddCommand(servicequota.New(prerunner))
	cmd.AddCommand(schemaregistry.New(cfg, prerunner, nil))
	cmd.AddCommand(secret.New(prerunner, flagResolver, secrets.NewPasswordProtectionPlugin()))
	cmd.AddCommand(shell.New(cmd))
	cmd.AddCommand(update.New(prerunner, ver, updateClient))
	cmd.AddCommand(version.New(prerunner, ver))

	changeDefaults(cmd, cfg)
	return cmd
}

func Execute(cmd *cobra.Command, args []string, cfg *v1.Config, ver *pversion.Version, isTest bool) error {
	if !cfg.DisablePlugins {
		if plugin, err := pplugin.FindPlugin(cmd, args); err != nil {
			return err
		} else if plugin != nil {
			pluginPath := plugin.Args[0]
			if strings.HasSuffix(pluginPath, ".sh") {
				dat, err := os.ReadFile(pluginPath)
				if err != nil {
					return err
				}
				if string(dat[:2]) != "#!" {
					shell := os.Getenv("SHELL")
					if shell == "" {
						utils.ErrPrintln(cmd, "SHELL env variable is empty")
					}
					shebang := []byte("#!" + shell + "\n")
					temp, err := os.CreateTemp(strings.TrimSuffix(pluginPath, filepath.Base(pluginPath)), plugin.Name)
					defer func() {
						err = temp.Close()
						if err != nil {
							fmt.Println(err)
						}
						err = os.Remove(temp.Name())
						if err != nil {
							fmt.Println(err)
						}
					}()
					if err != nil {
						fmt.Println(err)
						return err
					}
					if _, err := temp.Write(append(shebang, dat...)); err != nil {
						fmt.Println(err)
						return err
					}
					err = temp.Chmod(0755)
					if err != nil {
						fmt.Println(err)
						return err
					}
					plugin.Args[0] = temp.Name()
				}
			}
			return pplugin.ExecPlugin(plugin)
		}
	}
	// Usage collection is a wrapper around Execute() instead of a post-run function so we can collect the error status.
	u := usage.New(ver.Version)

	if !isTest && ver.IsReleased() {
		cmd.PersistentPostRun = u.Collect
	}

	err := cmd.Execute()
	errors.DisplaySuggestionsMessage(err, os.Stderr)

	if cfg.IsCloudLogin() && u.Command != nil && *(u.Command) != "" {
		ctx := cfg.Context()
		client := ccloudv2.NewClient(ctx.GetPlatformServer(), ver.UserAgent, isTest, ctx.GetAuthToken())

		u.Error = cliv1.PtrBool(err != nil)
		u.Report(client)
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

func getCloudClient(cfg *v1.Config, ccloudClientFactory pauth.CCloudClientFactory) *ccloud.Client {
	if cfg.IsCloudLogin() {
		return ccloudClientFactory.AnonHTTPClientFactory(pauth.CCloudURL)
	}
	return nil
}
