package iam

import (
	"github.com/spf13/cobra"

	ssov2 "github.com/confluentinc/ccloud-sdk-go-v2/sso/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *groupMappingCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a group mapping.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a group mapping named "DemoGroupMapping".`,
				Code: `confluent iam group-mapping create DemoGroupMapping --description "description of group mapping" --filter "\"demo\" in claims.group"`,
			},
		),
	}

	cmd.Flags().String("description", "", "Description of the group mapping.")
	cmd.Flags().String("filter", "", "CEL-compliant filter for this group mapping")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *groupMappingCommand) create(cmd *cobra.Command, args []string) error {
	name := args[0]

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		return err
	}

	newGroupMapping := ssov2.IamV2SsoGroupMapping{
		DisplayName: ssov2.PtrString(name),
		Description: ssov2.PtrString(description),
		Filter:      ssov2.PtrString(filter),
	}
	groupMapping, err := c.V2Client.CreateGroupMapping(newGroupMapping)
	if err != nil {
		return err
	}
	return printGroupMapping(cmd, groupMapping)
}
