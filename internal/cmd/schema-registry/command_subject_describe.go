package schemaregistry

import (
	"context"
	"fmt"

	"github.com/antihax/optional"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func (c *subjectCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <subject>",
		Short: "Describe subject versions.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Retrieve all versions registered under subject "payments" and its compatibility level.`,
				Code: fmt.Sprintf("%s schema-registry subject describe payments", version.CLIName),
			},
		),
	}

	cmd.Flags().BoolP("deleted", "D", false, "View the deleted schema.")
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *subjectCommand) describe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	return listSubjectVersions(cmd, args[0], srClient, ctx)
}

func listSubjectVersions(cmd *cobra.Command, subject string, srClient *srsdk.APIClient, ctx context.Context) error {
	deleted, err := cmd.Flags().GetBool("deleted")
	if err != nil {
		return err
	}

	listVersionsOpts := srsdk.ListVersionsOpts{Deleted: optional.NewBool(deleted)}
	versions, httpResp, err := srClient.DefaultApi.ListVersions(ctx, subject, &listVersionsOpts)
	if err != nil {
		return errors.CatchSchemaNotFoundError(err, httpResp)
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
