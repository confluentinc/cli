package cmd

import (
	"context"
	"github.com/confluentinc/cli/internal/cmd/connector"
	connector_catalog "github.com/confluentinc/cli/internal/cmd/connector-catalog"
	"net/http"
	"os"
	"runtime"

	"github.com/DABH/go-basher"
	"github.com/jonboulle/clockwork"
	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/mds-sdk-go"

	"github.com/confluentinc/cli/internal/cmd/apikey"
	"github.com/confluentinc/cli/internal/cmd/auth"
	"github.com/confluentinc/cli/internal/cmd/cluster"
	"github.com/confluentinc/cli/internal/cmd/completion"
	"github.com/confluentinc/cli/internal/cmd/config"
	"github.com/confluentinc/cli/internal/cmd/environment"
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
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	configs "github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/help"
	"github.com/confluentinc/cli/internal/pkg/io"
	"github.com/confluentinc/cli/internal/pkg/keystore"
	"github.com/confluentinc/cli/internal/pkg/log"
	pps1 "github.com/confluentinc/cli/internal/pkg/ps1"
	secrets "github.com/confluentinc/cli/internal/pkg/secret"
	versions "github.com/confluentinc/cli/internal/pkg/version"
)

type Command struct {
	*cobra.Command
	// @VisibleForTesting
	Analytics analytics.Client
	logger    *log.Logger
}

func NewConfluentCommand(cliName string, cfg *configs.Config, ver *versions.Version, logger *log.Logger, analytics analytics.Client) (*Command, error) {
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

	client := ccloud.NewClientWithJWT(context.Background(), cfg.AuthToken, &ccloud.Params{
		BaseURL: cfg.AuthURL, Logger: cfg.Logger, UserAgent: ver.UserAgent,
	})

	ch := &pcmd.ConfigHelper{Config: cfg, Client: client, Version: ver}
	fs := &io.RealFileSystem{}

	prerunner := &pcmd.PreRun{
		UpdateClient: updateClient,
		CLIName:      cliName,
		Version:      ver.Version,
		Logger:       logger,
		Config:       cfg,
		ConfigHelper: ch,
		Clock:        clockwork.NewRealClock(),
		Analytics:    analytics,
	}

	cli.PersistentPreRunE = prerunner.Anonymous()

	mdsConfig := mds.NewConfiguration()
	mdsConfig.BasePath = cfg.AuthURL
	mdsConfig.UserAgent = ver.UserAgent
	if cfg.Platforms[cfg.AuthURL] != nil {
		caCertPath := cfg.Platforms[cfg.AuthURL].CaCertPath
		if caCertPath != "" {
			// Try to load certs. On failure, warn, but don't error out because this may be an auth command, so there may
			// be a --ca-cert-path flag on the cmd line that'll fix whatever issue there is with the cert file in the config
			caCertFile, err := os.Open(caCertPath)
			if err == nil {
				defer caCertFile.Close()
				mdsConfig.HTTPClient, err = auth.SelfSignedCertClient(caCertFile, logger)
				if err != nil {
					logger.Warnf("Unable to load certificate from %s. %s. Resulting SSL errors will be fixed by logging in with the --ca-cert-path flag.", caCertPath, err.Error())
					mdsConfig.HTTPClient = auth.DefaultClient()
				}
			} else {
				logger.Warnf("Unable to load certificate from %s. %s. Resulting SSL errors will be fixed by logging in with the --ca-cert-path flag.", caCertPath, err.Error())
				mdsConfig.HTTPClient = auth.DefaultClient()
			}
		}
	}
	mdsClient := mds.NewAPIClient(mdsConfig)

	cli.Version = ver.Version
	cli.AddCommand(version.NewVersionCmd(prerunner, ver))

	conn := config.New(cfg, prerunner, analytics)
	conn.Hidden = true // The config/context feature isn't finished yet, so let's hide it
	cli.AddCommand(conn)

	cli.AddCommand(completion.NewCompletionCmd(cli, cliName))

	if !cfg.DisableUpdates {
		cli.AddCommand(update.New(cliName, cfg, ver, prompt, updateClient))
	}
	cli.AddCommand(auth.New(prerunner, cfg, logger, mdsClient, ver.UserAgent, analytics)...)

	resolver := &pcmd.FlagResolverImpl{Prompt: prompt, Out: os.Stdout}

	if cliName == "ccloud" {
		cmd, err := kafka.New(prerunner, cfg, logger.Named("kafka"), ver.ClientID, client.Kafka, ch)
		if err != nil {
			return nil, err
		}
		cli.AddCommand(cmd)
		cli.AddCommand(initcontext.New(prerunner, cfg, prompt, resolver, analytics))
		credType, err := cfg.CredentialType()
		if _, ok := err.(*errors.UnspecifiedCredentialError); ok {
			return nil, err
		}
		if credType == configs.APIKey {
			return &Command{Command: cli, Analytics: analytics, logger: logger}, nil
		}
		cli.AddCommand(ps1.NewPromptCmd(cfg, &pps1.Prompt{Config: cfg}, logger))
		ks := &keystore.ConfigKeyStore{Config: cfg, Helper: ch}
		cli.AddCommand(environment.New(prerunner, cfg, client.Account, cliName))
		cli.AddCommand(service_account.New(prerunner, cfg, client.User))
		cli.AddCommand(apikey.New(prerunner, cfg, client.APIKey, ch, ks))

		// Schema Registry
		// If srClient is nil, the function will look it up after prerunner verifies authentication. Exposed so tests can pass mocks
		sr := schema_registry.New(prerunner, cfg, client.SchemaRegistry, ch, nil, client.Metrics, logger)
		cli.AddCommand(sr)
		cli.AddCommand(ksql.New(prerunner, cfg, client.KSQL, client.Kafka, client.User, ch))
		cli.AddCommand(connector.New(prerunner, cfg, client.Connect, ch))
		cli.AddCommand(connector_catalog.New(prerunner, cfg, client.Connect, ch))
	} else if cliName == "confluent" {
		cli.AddCommand(iam.New(prerunner, cfg, mdsClient))

		metaClient := cluster.NewScopedIdService(&http.Client{}, ver.UserAgent, logger)
		cli.AddCommand(cluster.New(prerunner, cfg, metaClient))

		if runtime.GOOS != "windows" {
			bash, err := basher.NewContext("/bin/bash", false)
			if err != nil {
				return nil, err
			}
			shellRunner := &local.BashShellRunner{BasherContext: bash}
			cli.AddCommand(local.New(cli, prerunner, shellRunner, logger, fs))
		}

		cli.AddCommand(secret.New(prerunner, cfg, prompt, resolver, secrets.NewPasswordProtectionPlugin(logger)))
	}
	return &Command{Command: cli, Analytics: analytics, logger: logger}, nil
}

func (c *Command) Execute(args []string) error {
	c.Analytics.SetStartTime()
	c.Command.SetArgs(args)
	err := c.Command.Execute()
	if err != nil {
		analyticsError := c.Analytics.SendCommandFailed(err)
		if analyticsError != nil {
			c.logger.Debugf("segment analytics sending event failed: %s\n", analyticsError.Error())
		}
		return err
	}
	c.Analytics.CatchHelpCall(c.Command, args)
	analyticsError := c.Analytics.SendCommandSucceeded()
	if analyticsError != nil {
		c.logger.Debugf("segment analytics sending event failed: %s\n", analyticsError.Error())
	}
	return nil
}
