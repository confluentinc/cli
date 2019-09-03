package schema_registry

import (
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"
)

type updateCommand struct {
	*cobra.Command
	config   *config.Config
	ch       *pcmd.ConfigHelper
	srClient *srsdk.APIClient
}

// NewUpdateCommand returns the Cobra command for Schema Registry update.
func NewUpdateCommand(config *config.Config, ch *pcmd.ConfigHelper, srClient *srsdk.APIClient) *cobra.Command {
	updateCmd := &updateCommand{
		Command: &cobra.Command{
			Use:   "update",
			Short: "Update Schema Registry update.",
		},
		config:   config,
		ch:       ch,
		srClient: srClient,
	}
	updateCmd.init()
	return updateCmd.Command
}

func (c *updateCommand) init() {

	// Update
	cmd := &cobra.Command{
		Use:   "update <update> [--subject <subject>]",
		Short: "Update update for Schema Registry.",
		Example: `
Update  or Subject level update of schema registry.

::
		ccloud schema-registry update READWRITE
		ccloud schema-registry update update --subject subjectname READWRITE
`,
		RunE: c.update,
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)
}

func (c *updateCommand) update(cmd *cobra.Command, args []string) error {

	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}
	srClient, ctx, err := GetApiClient(c.srClient, c.ch)
	if err != nil {
		return err
	}

	if subject == "" {

		modeUpdate, _, err := srClient.DefaultApi.UpdateTopLevelMode(ctx, srsdk.ModeUpdateRequest{Mode: args[0]})
		if err != nil {
			return err
		}
		fmt.Println("Successfully updated Top Level update: " + modeUpdate.Mode)
	} else {
		modeUpdate, _, err := srClient.DefaultApi.UpdateMode(ctx, subject, srsdk.ModeUpdateRequest{Mode: args[0]})
		if err != nil {
			return err
		}
		fmt.Println("Successfully updated Subject level update: " + modeUpdate.Mode)
	}

	return nil
}
