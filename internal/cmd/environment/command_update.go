package environment

import (
	"github.com/spf13/cobra"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	nameconversions "github.com/confluentinc/cli/internal/pkg/name-conversions"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id|name>",
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
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	environment := orgv2.OrgV2Environment{DisplayName: orgv2.PtrString(name)}
	if environment, err = c.V2Client.UpdateOrgEnvironment(args[0], environment); err != nil {
		environmentId, err := nameconversions.ConvertEnvironmentNameToId(args[0], c.V2Client, false)
		if err != nil {
			return err
		}
		if environment, err = c.V2Client.UpdateOrgEnvironment(environmentId, environment); err != nil {
			return err
		}
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		IsCurrent: environment.GetId() == c.Context.GetCurrentEnvironment(),
		Id:        environment.GetId(),
		Name:      environment.GetDisplayName(),
	})

	return table.Print()
}
