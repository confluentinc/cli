package cmd

import (
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/DABH/go-basher"
	"github.com/jonboulle/clockwork"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/apikey"
	"github.com/confluentinc/cli/internal/cmd/auth"
	"github.com/confluentinc/cli/internal/cmd/cluster"
	"github.com/confluentinc/cli/internal/cmd/completion"
	"github.com/confluentinc/cli/internal/cmd/config"
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
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/help"
	"github.com/confluentinc/cli/internal/pkg/io"
	"github.com/confluentinc/cli/internal/pkg/log"
	pps1 "github.com/confluentinc/cli/internal/pkg/ps1"
	secrets "github.com/confluentinc/cli/internal/pkg/secret"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

type Command struct {
	*cobra.Command
	// @VisibleForTesting
	Analytics analytics.Client
	logger    *log.Logger
}

func NewConfluentCommand(cliName string, cfg *v3.Config, logger *log.Logger, ver *pversion.Version, analytics analytics.Client, netrcHandler *pauth.NetrcHandler) (*Command, error) {
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

	updateClient, err := update.NewClient(cliName, cfg.DisableUpdateCheck || cfg.DisableUpdates, logger)
	if err != nil {
		return nil, err
	}
	currCtx := cfg.Context()

	fs := &io.RealFileSystem{}

	resolver := &pcmd.FlagResolverImpl{Prompt: prompt, Out: os.Stdout}
	prerunner := &pcmd.PreRun{
		UpdateClient:       updateClient,
		CLIName:            cliName,
		Logger:             logger,
		Clock:              clockwork.NewRealClock(),
		FlagResolver:       resolver,
		Version:            ver,
		Analytics:          analytics,
		UpdateTokenHandler: pauth.NewUpdateTokenHandler(netrcHandler),
	}
	_ = pcmd.NewAnonymousCLICommand(cli, cfg, prerunner) // Add to correctly set prerunners. TODO: Check if really needed.
	command := &Command{Command: cli, Analytics: analytics, logger: logger}

	cli.Version = ver.Version
	cli.AddCommand(version.NewVersionCmd(prerunner, ver))

	conn := config.New(cfg, prerunner, analytics)
	conn.Hidden = true // The config/context feature isn't finished yet, so let's hide it
	cli.AddCommand(conn)

	cli.AddCommand(completion.NewCompletionCmd(cli, cliName))

	if !cfg.DisableUpdates {
		cli.AddCommand(update.New(cliName, cfg, ver, prompt, updateClient, analytics))
	}
	cli.AddCommand(auth.New(prerunner, cfg, logger, ver.UserAgent, analytics, netrcHandler)...)

	if cliName == "ccloud" {
		cmd := kafka.New(prerunner, cfg, logger.Named("kafka"), ver.ClientID)
		cli.AddCommand(cmd)
		cli.AddCommand(feedback.NewFeedbackCmd(prerunner, cfg, analytics))
		cli.AddCommand(initcontext.New(prerunner, cfg, prompt, resolver, analytics))
		if currCtx != nil && currCtx.Credential != nil && currCtx.Credential.CredentialType == v2.APIKey {
			return command, nil
		}
		cli.AddCommand(ps1.NewPromptCmd(cfg, &pps1.Prompt{Config: cfg}, logger))
		cli.AddCommand(environment.New(prerunner, cfg, cliName))
		cli.AddCommand(service_account.New(prerunner, cfg))
		// Keystore exposed so tests can pass mocks.
		cli.AddCommand(apikey.New(prerunner, cfg, nil, resolver))

		// Schema Registry
		// If srClient is nil, the function will look it up after prerunner verifies authentication. Exposed so tests can pass mocks
		sr := schema_registry.New(prerunner, cfg, nil, logger)
		cli.AddCommand(sr)
		cli.AddCommand(ksql.New(prerunner, cfg))
		cli.AddCommand(connector.New(prerunner, cfg))
		cli.AddCommand(connector_catalog.New(prerunner, cfg))

		conn = ksql.New(prerunner, cfg)
		conn.Hidden = true // The ksql feature isn't finished yet, so let's hide it
		cli.AddCommand(conn)

		//conn = connect.New(prerunner, cfg, connects.New(client, logger))
		//conn.Hidden = true // The connect feature isn't finished yet, so let's hide it
		//cli.AddCommand(conn)
	} else if cliName == "confluent" {
		cli.AddCommand(iam.New(prerunner, cfg))

		metaClient := cluster.NewScopedIdService(&http.Client{}, ver.UserAgent, logger)
		cli.AddCommand(cluster.New(prerunner, cfg, metaClient))

		if runtime.GOOS != "windows" {
			bash, err := basher.NewContext("/bin/bash", false)
			if err != nil {
				return nil, err
			}
			shellRunner := &local.BashShellRunner{BasherContext: bash}
			cli.AddCommand(local.New(cli, prerunner, shellRunner, logger, fs, cfg))
		}

		command := local.NewCommand(prerunner, cfg)
		command.Hidden = true // WIP
		cli.AddCommand(command)

		cli.AddCommand(secret.New(prompt, resolver, secrets.NewPasswordProtectionPlugin(logger)))
	}
	return command, nil
}

func (c *Command) Execute(cliName string, args []string) error {
	c.Analytics.SetStartTime()
	c.Command.SetArgs(args)

	err := c.Command.Execute()
	analyticsError := c.Analytics.SendCommandAnalytics(c.Command, args, err)
	if analyticsError != nil {
		c.logger.Debugf("segment analytics sending event failed: %s\n", analyticsError.Error())
	}

	if cliName == "ccloud" && isHumanReadable(args) {
		failed := err != nil
		c.sendFeedbackNudge(failed, args)
	}

	return err
}

func isHumanReadable(args []string) bool {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "-o" {
			return args[i+1] == "human"
		}
	}
	return true
}

func (c *Command) sendFeedbackNudge(failed bool, args []string) {
	feedbackNudge := "\nDid you know you can use the \"ccloud feedback\" command to send the team feedback?\nLet us know if the ccloud CLI is meeting your needs, or what we can do to improve it."

	if failed {
		c.PrintErrln(feedbackNudge)
		return
	}

	feedbackNudgeCmds := []string{
		"api-key create", "api-key delete",
		"connector create", "connector delete",
		"environment create", "environment delete",
		"kafka acl create", "kafka acl delete",
		"kafka cluster create", "kafka cluster delete",
		"kafka topic create", "kafka topic delete",
		"ksql app create", "ksql app delete",
		"schema-registry schema create", "schema-registry schema delete",
		"service-account create", "service-account delete",
	}

	cmd := strings.Join(args, " ")
	for _, cmdPrefix := range feedbackNudgeCmds {
		if strings.HasPrefix(cmd, cmdPrefix) {
			c.PrintErrln(feedbackNudge)
			return
		}
	}
}
