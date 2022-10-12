package streamshare

import (
	"fmt"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.optOut(cmd, args, form.NewPrompt(os.Stdin))
		},
	}
}

func (c *command) optOut(cmd *cobra.Command, _ []string, prompt *form.RealPrompt) error {
	isDeleteConfirmed, err := confirmOptOut(cmd, prompt)
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

func confirmOptOut(cmd *cobra.Command, prompt form.Prompt) (bool, error) {
	f := form.New(
		form.Field{
			ID: "confirmation",
			Prompt: fmt.Sprintf("Are you sure you want to disable Stream Sharing for your organization? " +
				"Existing shares in your organization will not be accessible if Stream Sharing is disabled."),
			IsYesOrNo: true,
		},
	)
	if err := f.Prompt(cmd, prompt); err != nil {
		return false, errors.New(errors.FailedToReadOptOutConfirmationErrorMsg)
	}
	return f.Responses["confirmation"].(bool), nil
}
