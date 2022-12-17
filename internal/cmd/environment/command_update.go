package environment

import (
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an existing Confluent Cloud environment.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
	}

	cmd.Flags().String("name", "", "New name for Confluent Cloud environment.")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	id := args[0]

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	updateEnvironment := orgv2.OrgV2Environment{DisplayName: orgv2.PtrString(name)}
	_, httpResp, err := c.V2Client.UpdateOrgEnvironment(id, updateEnvironment)
	if err != nil {
		return errors.CatchOrgV2ResourceNotFoundError(err, resource.Environment, httpResp)
	}

	utils.ErrPrintf(cmd, errors.UpdateSuccessMsg, "name", "environment", id, name)
	return nil
}
