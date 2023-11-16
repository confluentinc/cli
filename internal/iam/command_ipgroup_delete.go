package iam

import (
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"
	"strings"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *ipGroupCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete an IP group.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete IP group "ipg-12345":`,
				Code: "confluent iam ip-group delete ipg-12345",
			},
		),
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *ipGroupCommand) delete(cmd *cobra.Command, args []string) error {
	err := c.V2Client.DeleteIamIpGroup(args[0])

	// Unique error message for deleting an IP group that has an IP filter bound to it
	if strings.Contains(err.Error(), "related IP filters") {
		return errors.NewErrorWithSuggestions(err.Error(),
			"List IP filters with `confluent iam ip-filter list`")
	}
	if err != nil {
		return resource.ResourcesNotFoundError(cmd, resource.IPGroup, args[0])
	}

	output.Printf(c.Config.EnableColor, "Deleted IP group \"%s\"\n", args[0])
	return nil
}
