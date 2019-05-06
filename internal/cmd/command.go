package cmd

import (
	"context"
	"os"

	"github.com/DABH/go-basher"
	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"

	"github.com/confluentinc/cli/internal/cmd/apikey"
	"github.com/confluentinc/cli/internal/cmd/auth"
	"github.com/confluentinc/cli/internal/cmd/completion"
	"github.com/confluentinc/cli/internal/cmd/config"
	"github.com/confluentinc/cli/internal/cmd/environment"
	"github.com/confluentinc/cli/internal/cmd/kafka"
	"github.com/confluentinc/cli/internal/cmd/ksql"
	"github.com/confluentinc/cli/internal/cmd/local"
	service_account "github.com/confluentinc/cli/internal/cmd/service-account"
	"github.com/confluentinc/cli/internal/cmd/update"
	"github.com/confluentinc/cli/internal/cmd/version"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	configs "github.com/confluentinc/cli/internal/pkg/config"
	keystore "github.com/confluentinc/cli/internal/pkg/keystore"
	"github.com/confluentinc/cli/internal/pkg/log"
	apikeys "github.com/confluentinc/cli/internal/pkg/sdk/apikey"
	//connects "github.com/confluentinc/cli/internal/pkg/sdk/connect"
	environments "github.com/confluentinc/cli/internal/pkg/sdk/environment"
	kafkas "github.com/confluentinc/cli/internal/pkg/sdk/kafka"
	ksqls "github.com/confluentinc/cli/internal/pkg/sdk/ksql"
	users "github.com/confluentinc/cli/internal/pkg/sdk/user"
	versions "github.com/confluentinc/cli/internal/pkg/version"
)

func NewConfluentCommand(cliName string, cfg *configs.Config, ver *versions.Version, logger *log.Logger) (*cobra.Command, error) {
	cli := &cobra.Command{
		Use:               cliName,
		Version:           ver.Version,
		DisableAutoGenTag: true,
	}
	if cliName == "ccloud" {
		cli.Short = "Confluent Cloud CLI"
		cli.Long = "Manage your Confluent Cloud"
	} else {
		cli.Short = "Confluent CLI"
		cli.Long = "Manage your Confluent Platform"
	}
	cli.PersistentFlags().CountP("verbose", "v",
		"increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace)")

	prompt := pcmd.NewPrompt(os.Stdin)

	updateClient, err := update.NewClient(cliName, logger)
	if err != nil {
		return nil, err
	}

	prerunner := &pcmd.PreRun{
		UpdateClient: updateClient,
		CLIName:      cliName,
		Version:      ver.Version,
		Logger:       logger,
		Config:       cfg,
	}

	cli.PersistentPreRunE = prerunner.Anonymous()

	client := ccloud.NewClientWithJWT(context.Background(), cfg.AuthToken, cfg.AuthURL, cfg.Logger)

	cli.Version = ver.Version
	cli.AddCommand(version.NewVersionCmd(prerunner, ver))

	conn := config.New(cfg)
	conn.Hidden = true // The config/context feature isn't finished yet, so let's hide it
	cli.AddCommand(conn)

	cli.AddCommand(completion.NewCompletionCmd(cli, cliName))
	cli.AddCommand(update.New(cliName, cfg, ver, prompt, updateClient))

	cli.AddCommand(auth.New(prerunner, cfg)...)

	if cliName == "ccloud" {
		kafkaClient := kafkas.New(client, logger)
		userClient := users.New(client, logger)
		ch := &pcmd.ConfigHelper{Config: cfg, Client: client}
		ks := &keystore.ConfigKeyStore{Config: cfg, Helper: ch}
		cli.AddCommand(environment.New(prerunner, cfg, environments.New(client, logger), cliName))
		cli.AddCommand(service_account.New(prerunner, cfg, userClient))
		cli.AddCommand(apikey.New(prerunner, cfg, apikeys.New(client, logger), ch, ks))
		cli.AddCommand(kafka.New(prerunner, cfg, kafkaClient, ch))

		conn = ksql.New(prerunner, cfg, ksqls.New(client, logger), kafkaClient, userClient, ch)
		conn.Hidden = true // The ksql feature isn't finished yet, so let's hide it
		cli.AddCommand(conn)

		//conn = connect.New(prerunner, cfg, connects.New(client, logger))
		//conn.Hidden = true // The connect feature isn't finished yet, so let's hide it
		//cli.AddCommand(conn)
	} else if cliName == "confluent" {
		bash, err := basher.NewContext("/bin/bash", false)
		if err != nil {
			return nil, err
		}
		shellRunner := local.BashShellRunner{BasherContext: bash}
		cli.AddCommand(local.New(prerunner, &shellRunner))
	}

	return cli, nil
}
