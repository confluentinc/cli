package stream_share

import (
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	prerunner       pcmd.PreRunner
	analyticsClient analytics.Client
}

// New returns the default command object to perform operations on stream share.
func New(prerunner pcmd.PreRunner, analyticsClient analytics.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "stream-share",
		Short:       "Manage stream share.",
		Long:        "Create and redeem shared token for a stream share.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &command{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
		prerunner:               prerunner,
		analyticsClient:         analyticsClient,
	}
	c.init()

	return c.Command
}

func (c *command) init() {
	c.AddCommand(NewSharedTokenCommand(c.prerunner, c.analyticsClient).Command)

	deactivateCommand := &cobra.Command{
		Use:   "deactivate",
		Short: "Deactivate a stream share.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.deactivate),
	}
	deactivateCommand.Flags().String("id", "", "The ID of the stream share to deactivate.")
	_ = deactivateCommand.MarkFlagRequired("id")
	c.AddCommand(deactivateCommand)
}

func (c *command) deactivate(cmd *cobra.Command, _ []string) error {
	id, err := cmd.Flags().GetString("id")
	if err != nil {
		return err
	} else if id == "" {
		return errors.New(errors.StreamShareIdEmptyErrorMsg)
	}

	err = c.Client.StreamShare.DeactivateStreamShare(id)

	if err != nil {
		return err
	}

	utils.Printf(cmd, "Stream share \"%s\" deactivated successfully.\n", id)

	return nil

}
