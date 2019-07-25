package schema_registry

import (
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"
)


type modeCommand struct {

	prerunner pcmd.PreRunner
	*cobra.Command
	config   *config.Config
	ch       *pcmd.ConfigHelper
	srClient *srsdk.APIClient
}

// NewModeCommand returns the Cobra command for Schema Registry mode.
func NewModeCommand(config *config.Config, ch *pcmd.ConfigHelper, srClient *srsdk.APIClient) *cobra.Command {
	compatCmd := &modeCommand{
		Command: &cobra.Command{
			Use:   "mode",
			Short: "Manage Schema Registry compatibility.",
		},
		config:   config,
		ch:       ch,
		srClient: srClient,
	}
	compatCmd.init()
	return compatCmd.Command
}


func (c *modeCommand) init() {

	// Update
	cmd := &cobra.Command{
		Use:   "update <mode> [--subject <subject>]",
		Short: "Update mode for Schema Registry.",
		Example: `
Update mode level for the specified subject to READWRITE, READONLY or IMPORT.

::
      ccloud schema-registry mode update IMPORT
`,
		RunE: c.update,
		Args: cobra.ExactArgs(1),
	}
	c.AddCommand(cmd)
}


func (c *modeCommand) update(cmd *cobra.Command, args []string) error {

	srClient, ctx, err := GetApiClient(c.srClient, c.ch)
	if err != nil {
		return err
	}

	updatedMode, _, err := srClient.DefaultApi.UpdateMode(ctx, args[0])
	if err != nil {
		return err
	}

	fmt.Println("Successfully updated to new mode: "+ updatedMode.Mode)
	return nil
}
