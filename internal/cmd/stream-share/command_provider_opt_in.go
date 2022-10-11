package streamshare

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newOptInCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "opt-in",
		Short: "Opt in to Stream Sharing.",
		RunE:  c.optIn,
	}
}

func (c *command) optIn(cmd *cobra.Command, _ []string) error {
	_, httpResp, err := c.V2Client.OptInOrOut(true)
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}

	utils.Print(cmd, errors.OptInMsg)
	return nil
}
