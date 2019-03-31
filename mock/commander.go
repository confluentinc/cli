package mock

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/terminal"
)

type Commander struct {
	Prompt terminal.Prompt
}

func (c *Commander) Anonymous() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if c.Prompt != nil {
			c.Prompt.SetOutput(cmd.OutOrStdout())
		}
		return nil
	}
}

func (c *Commander) Authenticated() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if c.Prompt != nil {
			c.Prompt.SetOutput(cmd.OutOrStdout())
		}
		return nil
	}
}
