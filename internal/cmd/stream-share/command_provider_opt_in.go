package streamshare

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newOptInCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "opt-in",
		Short: "Opt in to stream sharing.",
		RunE:  c.optIn,
	}
}

func (c *command) optIn(_ *cobra.Command, _ []string) error {
	if _, err := c.V2Client.StreamShareOptInOrOut(true); err != nil {
		return err
	}

	output.Println(errors.OptInMsg)
	return nil
}
