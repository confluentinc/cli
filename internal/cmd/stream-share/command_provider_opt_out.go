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

func (c *command) optOut(_ *cobra.Command, _ []string) error {
	isDeleteConfirmed, err := confirmOptOut()
	if err != nil {
		return err
	}
	if !isDeleteConfirmed {
		utils.Println("Operation terminated.")
		return nil
	}

	if _, err := c.V2Client.StreamShareOptInOrOut(false); err != nil {
		return err
	}

	utils.Println(errors.OptOutMsg)
	return nil
}
