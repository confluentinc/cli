package streamshare

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/output"
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
		output.Println("Operation terminated.")
		return nil
	}

	if _, err := c.V2Client.StreamShareOptInOrOut(false); err != nil {
		return err
	}

	output.Println("Successfully opted out of Stream Sharing.")
	return nil
}
