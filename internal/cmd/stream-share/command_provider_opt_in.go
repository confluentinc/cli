package streamshare

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newOptInCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "opt-in",
		Short: "Opt in to stream sharing.",
		RunE:  c.optIn,
	}
}

func (c *command) optIn(cmd *cobra.Command, _ []string) error {
	_, err := c.V2Client.StreamShareOptInOrOut(true)
	if err != nil {
		return err
	}

	utils.Print(cmd, errors.OptInMsg)
	return nil
}
