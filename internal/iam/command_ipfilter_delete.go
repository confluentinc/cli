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

func (c *ipFilterCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete an IP filter.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete IP filter "ipf-12345":`,
				Code: "confluent iam ip-filter delete ipf-12345",
			},
		),
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *ipFilterCommand) delete(cmd *cobra.Command, args []string) error {
	err := c.V2Client.DeleteIamIpFilter(args[0])

	/*
	 * Unique error message for deleting an IP Filter that would lock out the user.
	 * Splits the error message into its two components of the error and the suggestion.
	 */
	if strings.Contains(err.Error(), "lock out") {
		errorMessageIndex := strings.Index(err.Error(), "Please")
		return errors.NewErrorWithSuggestions(err.Error()[:errorMessageIndex-1],
			err.Error()[errorMessageIndex:])
	}
	if err != nil {
		return resource.ResourcesNotFoundError(cmd, resource.IPFilter, args[0])
	}

	output.Printf(c.Config.EnableColor, "Deleted IP filter \"%s\"\n", args[0])
	return nil
}
