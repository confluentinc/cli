package iam

import (
	"github.com/confluentinc/ccloud-sdk-go-v2/sso/v2"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type groupMappingCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type groupMappingOut struct {
	Id          string `human:"ID" serialized:"id"`
	Name        string `human:"Name" serialized:"name"`
	Description string `human:"Description" serialized:"description"`
	Filter      string `human:"Filter" serialized:"filter"`
	Principal   string `human:"Principal" serialized:"principal"`
	State       string `human:"State" serialized:"state"`
}

func newGroupMappingCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "group-mapping",
		Short:       "Manage SSO group mappings.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &groupMappingCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func printGroupMapping(cmd *cobra.Command, groupMapping sso.IamV2SsoGroupMapping) error {
	table := output.NewTable(cmd)
	table.Add(&groupMappingOut{
		Id:          groupMapping.GetId(),
		Name:        groupMapping.GetDisplayName(),
		Description: groupMapping.GetDescription(),
		Filter:      groupMapping.GetFilter(),
		Principal:   groupMapping.GetPrincipal(),
		State:       groupMapping.GetState(),
	})
	return table.Print()
}

func (c *groupMappingCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteGroupMappings(c.V2Client)
}
