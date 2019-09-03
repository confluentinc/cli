package schema_registry

import (
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"
)

type modeCommand struct {
	*cobra.Command
	config   *config.Config
	ch       *pcmd.ConfigHelper
	srClient *srsdk.APIClient
}

// NewModeCommand returns the Cobra command for Schema Registry mode.
func NewModeCommand(config *config.Config, ch *pcmd.ConfigHelper, srClient *srsdk.APIClient) *cobra.Command {
	modeCmd := &modeCommand{
		Command: &cobra.Command{
			Use:   "mode",
			Short: "Update Schema Registry mode.",
		},
		config:   config,
		ch:       ch,
		srClient: srClient,
	}
	modeCmd.init()
	return modeCmd.Command
}

func (c *modeCommand) init() {

	// Update
	cmd := &cobra.Command{
		Use:   "update <mode> [--subject <subject>]",
		Short: "Update mode for Schema Registry.",
		Example: `
Update Top level mode or Subject level mode of schema registry.

::
		ccloud schema-registry mode update READWRITE
		ccloud schema-registry mode update --subject subjectname READWRITE
`,
		RunE: c.update,
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)
}

func (c *modeCommand) update(cmd *cobra.Command, args []string) error {

	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}
	srClient, ctx, err := GetApiClient(c.srClient, c.ch)
	if err != nil {
		return err
	}

	if subject == "" {

		updatedMode, _, err := srClient.DefaultApi.UpdateTopLevelMode(ctx, srsdk.ModeUpdateRequest{Mode: args[0]})
		if err != nil {
			return err
		}
		fmt.Println("Successfully updated Top Level Mode: " + updatedMode.Mode)
	} else {
		updatedMode, _, err := srClient.DefaultApi.UpdateMode(ctx, subject, srsdk.ModeUpdateRequest{Mode: args[0]})
		if err != nil {
			return err
		}
		fmt.Println("Successfully updated Subject level Mode: " + updatedMode.Mode)
	}

	return nil
}
