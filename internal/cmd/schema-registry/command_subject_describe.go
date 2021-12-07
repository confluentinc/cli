package schemaregistry

import (
	"fmt"

	"github.com/antihax/optional"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func (c *subjectCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <subject>",
		Short: "Describe subject versions and compatibility.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.describe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Retrieve all versions registered under a given subject and its compatibility level.",
				Code: fmt.Sprintf("%s schema-registry subject describe <subject-name>", version.CLIName),
			},
		),
	}

	cmd.Flags().BoolP("deleted", "D", false, "View the deleted schema.")
	output.AddFlag(cmd)

	return cmd
}

func (c *subjectCommand) describe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	deleted, err := cmd.Flags().GetBool("deleted")
	if err != nil {
		return err
	}

	listVersionsOpts := srsdk.ListVersionsOpts{Deleted: optional.NewBool(deleted)}
	versions, _, err := srClient.DefaultApi.ListVersions(ctx, args[0], &listVersionsOpts)
	if err != nil {
		return err
	}

	outputOption, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	if outputOption == output.Human.String() {
		printVersions(versions)
	} else {
		structuredOutput := &struct{ Version []int32 }{versions}
		fields := []string{"Version"}
		structuredRenames := map[string]string{"Version": "version"}
		return output.DescribeObject(cmd, structuredOutput, fields, map[string]string{}, structuredRenames)
	}

	return nil
}
