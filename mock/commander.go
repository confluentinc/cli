package mock

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type Commander struct {
	FlagResolver cmd.FlagResolver
}

var _ cmd.PreRunner = (*Commander)(nil)

func NewPreRunnerMock() cmd.PreRunner {
	flagResolverMock := &cmd.FlagResolverImpl{
		Prompt: &Prompt{},
		Out:    os.Stdout,
	}
	return &Commander{FlagResolver: flagResolverMock}
}

func (c *Commander) Anonymous(cfg *config.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}

func (c *Commander) Authenticated(cfg *config.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := c.FlagResolver.ResolveFlags(cmd, cfg)
		if err != nil {
			panic(err)
		}
		return nil
	}
}

func (c *Commander) HasAPIKey(cfg *config.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := c.FlagResolver.ResolveFlags(cmd, cfg)
		if err != nil {
			panic(err)
		}
		return nil
	}
}
