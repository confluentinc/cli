package environment

import (
	"github.com/spf13/cobra"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
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
	cmd.Flags().String("stream-governance-package", "",
		"Stream Governance package (\"ESSENTIALS\" or \"ADVANCED\")")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsOneRequired("name", "stream-governance-package")

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	sgPackage, err := cmd.Flags().GetString("stream-governance-package")
	if err != nil {
		return err
	}

	environment := orgv2.OrgV2Environment{}
	if name != "" {
		environment.SetDisplayName(name)
	}
	if sgPackage != "" {
		sgConfig := orgv2.NewOrgV2StreamGovernanceConfig()
		sgConfig.SetPackage(sgPackage)
		environment.SetStreamGovernanceConfig(*sgConfig)
	}

	environment, err = c.V2Client.UpdateOrgEnvironment(args[0], environment)
	if err != nil {
		return err
	}
	newSgConfig := environment.GetStreamGovernanceConfig()

	table := output.NewTable(cmd)
	table.Add(&out{
		IsCurrent: environment.GetId() == c.Context.GetCurrentEnvironment(),
		Id:        environment.GetId(),
		Name:      environment.GetDisplayName(),
		SGPackage: newSgConfig.GetPackage(),
	})
	return table.Print()
}
