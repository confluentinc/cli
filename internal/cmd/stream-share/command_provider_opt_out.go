package streamshare

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newOptOutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "opt-out",
		Short: "Opt out of stream sharing.",
		RunE:  c.optOut,
	}
}

func (c *command) optOut(cmd *cobra.Command, _ []string) error {
	isDeleteConfirmed, err := confirmOptOut(cmd)
	if err != nil {
		return err
	}

	if !isDeleteConfirmed {
		return nil
	}

	_, err = c.V2Client.StreamShareOptInOrOut(false)
	if err != nil {
		return err
	}

	utils.Print(cmd, errors.OptOutMsg)
	return nil
}
