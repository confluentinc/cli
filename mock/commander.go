package mock

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type Commander struct{}

var _ cmd.PreRunner = (*Commander)(nil)

func (c *Commander) Anonymous() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}

func (c *Commander) Authenticated() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}

func (c *Commander) HasAPIKey() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}

func (c *Commander) Context() *config.Context {
	return config.AuthenticatedConfigMock().Context()
}
