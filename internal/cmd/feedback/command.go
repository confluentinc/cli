package feedback

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/log"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

func NewFeedbackCmd(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	prompt := form.NewPrompt(os.Stdin)
	return NewFeedbackCmdWithPrompt(cfg, prerunner, prompt)
}

func NewFeedbackCmdWithPrompt(cfg *v1.Config, prerunner pcmd.PreRunner, prompt form.Prompt) *cobra.Command {
	cmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "feedback",
			Short: "Submit feedback about the " + pversion.FullCLIName + ".",
			RunE: func(cmd *cobra.Command, _ []string) error {
				msg, err := getFeedback(prompt)
				if err != nil {
					return err
				}
				if len(msg) > 0 {
					if err := sendFeedback(cfg, cmd, msg); err != nil {
						return err
					}
					fmt.Println("Thanks for your feedback.")
				}
				return nil
			},
			Args: cobra.NoArgs,
		},
		prerunner)

	return cmd.Command
}

func getFeedback(prompt form.Prompt) (string, error) {
	fmt.Println("Please confirm that your feedback will not contain any " +
		"sensitive information such as API-keys, unencrypted tokens, etc.")
	fmt.Print("Type \"y\" to confirm or anything else to exit: ")
	ack, err := prompt.ReadLine()
	if err != nil {
		return "", err
	}
	if ack != "y" {
		return "", err
	}
	fmt.Println("Enter feedback: ")
	msg, err := prompt.ReadLine()
	if err != nil {
		return "", err
	}
	msg = strings.TrimLeft(msg, " ")
	msg = strings.TrimRight(msg, " \n")
	return msg, err
}

func sendFeedback(cfg *v1.Config, cmd *cobra.Command, msg string) error {
	feedback := cliv1.CliV1Feedback{Content: cliv1.PtrString(msg)}
	unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
	if err != nil {
		return err
	}
	if err := cfg.GetCloudClientV2(unsafeTrace).CreateCliFeedback(feedback); err != nil {
		log.CliLogger.Warnf("Failed to report CLI feedback: %v", err)
	}
	return nil
}
