package environment

import (
	"github.com/spf13/cobra"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
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
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("name"))

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	id := args[0]

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	updateEnvironment := orgv2.OrgV2Environment{DisplayName: orgv2.PtrString(name)}
	environment, httpResp, err := c.V2Client.UpdateOrgEnvironment(id, updateEnvironment)
	if err != nil {
		return errors.CatchOrgV2ResourceNotFoundError(err, resource.Environment, httpResp)
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		IsCurrent: environment.GetId() == c.Context.GetCurrentEnvironment(),
		Id:        environment.GetId(),
		Name:      environment.GetDisplayName(),
	})
	return table.Print()
}
