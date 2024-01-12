package environment

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new Confluent Cloud environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
	}

	c.addStreamGovernancePackageFlag(cmd, "essentials")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	governancePackage, err := cmd.Flags().GetString("governance-package")
	if err != nil {
		return err
	}

	environment := orgv2.OrgV2Environment{DisplayName: orgv2.PtrString(args[0])}
	if governancePackage != "" {
		environment.SetStreamGovernanceConfig(orgv2.OrgV2StreamGovernanceConfig{
			Package: orgv2.PtrString(strings.ToUpper(governancePackage)),
		})
	}

	environment, err = c.V2Client.CreateOrgEnvironment(environment)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		IsCurrent:               environment.GetId() == c.Context.GetCurrentEnvironment(),
		Id:                      environment.GetId(),
		Name:                    environment.GetDisplayName(),
		StreamGovernancePackage: environment.StreamGovernanceConfig.GetPackage(),
	})
	if err := table.Print(); err != nil {
		return err
	}

	c.Context.AddEnvironment(environment.GetId())
	_ = c.Config.Save()

	return nil
}

func (c *command) addStreamGovernancePackageFlag(cmd *cobra.Command, defaultValue string) {
	values := utils.ArrayToCommaDelimitedString([]string{"essentials", "advanced"}, "or")
	cmd.Flags().String("governance-package", defaultValue, fmt.Sprintf("Specify the Stream Governance package as %s.", values))
}
