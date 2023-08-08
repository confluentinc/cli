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
				Text: `Create a group mapping named "demo-group-mapping".`,
				Code: `confluent iam group-mapping create demo-group-mapping --description "new description" --filter "\"demo\" in claims.group"`,
			},
		),
	}

	cmd.Flags().String("description", "", "Description of the group mapping.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddFilterFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *groupMappingCommand) create(cmd *cobra.Command, args []string) error {
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		return err
	}

	createGroupMapping := ssov2.IamV2SsoGroupMapping{
		DisplayName: ssov2.PtrString(args[0]),
		Description: ssov2.PtrString(description),
		Filter:      ssov2.PtrString(filter),
	}
	groupMapping, err := c.V2Client.CreateGroupMapping(createGroupMapping)
	if err != nil {
		return err
	}
	return printGroupMapping(cmd, groupMapping)
}
