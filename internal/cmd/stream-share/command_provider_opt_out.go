package streamshare

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
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
		utils.Println(cmd, "Operation terminated.")
		return nil
	}

	_, err = c.V2Client.OptInOrOut(false)
	if err != nil {
		return err
	}

	utils.Print(cmd, errors.OptOutMsg)
	return nil
}

func confirmOptOut(cmd *cobra.Command) (bool, error) {
	f := form.New(
		form.Field{
			ID: "confirmation",
			Prompt: "Are you sure you want to disable Stream Sharing for your organization? " +
				"Existing shares in your organization will not be accessible if Stream Sharing is disabled.",
			IsYesOrNo: true,
		},
	)
	if err := f.Prompt(cmd, form.NewPrompt(os.Stdin)); err != nil {
		return false, errors.New(errors.FailedToReadOptOutConfirmationErrorMsg)
	}
	return f.Responses["confirmation"].(bool), nil
}
