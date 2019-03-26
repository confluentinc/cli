package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/cli/internal/cmd/apikey"
	"github.com/confluentinc/cli/internal/cmd/auth"
	"github.com/confluentinc/cli/internal/cmd/completion"
	"github.com/confluentinc/cli/internal/cmd/config"
	"github.com/confluentinc/cli/internal/cmd/environment"
	"github.com/confluentinc/cli/internal/cmd/kafka"
	"github.com/confluentinc/cli/internal/cmd/ksql"
	"github.com/confluentinc/cli/internal/cmd/service-account"
	"github.com/confluentinc/cli/internal/cmd/version"
	configs "github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	apikeys "github.com/confluentinc/cli/internal/pkg/sdk/apikey"
	//connects "github.com/confluentinc/cli/pkg/sdk/connect"
	environments "github.com/confluentinc/cli/internal/pkg/sdk/environment"
	kafkas "github.com/confluentinc/cli/internal/pkg/sdk/kafka"
	ksqls "github.com/confluentinc/cli/internal/pkg/sdk/ksql"
	users "github.com/confluentinc/cli/internal/pkg/sdk/user"
	"github.com/confluentinc/cli/internal/pkg/terminal"
	versions "github.com/confluentinc/cli/internal/pkg/version"
)

const cliName = "ccloud"

func NewConfluentCommand(cfg *configs.Config, ver *versions.Version, logger *log.Logger) *cobra.Command {
	cli := &cobra.Command{
		Use:   cliName,
		Short: "Welcome to the Confluent Cloud CLI",
	}
	cli.PersistentFlags().CountP("verbose", "v", "increase output verbosity")
	cli.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := log.SetLoggingVerbosity(cmd, logger); err != nil {
			return errors.HandleCommon(err, cmd)
		}
		return nil
	}

	prompt := terminal.NewPrompt(os.Stdin)

	client := ccloud.NewClientWithJWT(context.Background(), cfg.AuthToken, cfg.AuthURL, cfg.Logger)

	cli.Version = ver.Version
	cli.AddCommand(version.NewVersionCmd(ver, prompt))

	conn := config.New(cfg)
	conn.Hidden = true // The config/context feature isn't finished yet, so let's hide it
	cli.AddCommand(conn)

	conn, err := completion.NewCompletionCmd(cli, prompt, cliName)
	if err != nil {
		logger.Log("msg", err)
	} else {
		cli.AddCommand(conn)
	}

	cli.AddCommand(auth.New(cfg)...)
	cli.AddCommand(environment.New(cfg, environments.New(client, logger)))
	cli.AddCommand(service_account.New(cfg, users.New(client, logger)))
	cli.AddCommand(apikey.New(cfg, apikeys.New(client, logger)))
	cli.AddCommand(kafka.New(cfg, kafkas.New(client, logger)))

	conn = ksql.New(cfg, ksqls.New(client, logger))
	conn.Hidden = true // The ksql feature isn't finished yet, so let's hide it
	cli.AddCommand(conn)

	//conn = connect.New(cfg, connects.New(client, logger))
	//conn.Hidden = true // The connect feature isn't finished yet, so let's hide it
	//cli.AddCommand(conn)

	return cli
}
