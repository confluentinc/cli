package schemaregistry

import (
	"context"
	"fmt"
	"net/http"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"
)

func (c *configCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe the config of a subject, or at global level.",
		Args:  cobra.MaximumNArgs(0),
		RunE:  pcmd.NewCLIRunE(c.describe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the configuration of subject `payments`.",
				Code: fmt.Sprintf("%s schema-registry config describe --subject payments", pversion.CLIName),
			},
			examples.Example{
				Text: "Describe the global configuration.",
				Code: fmt.Sprintf("%s schema-registry config describe", pversion.CLIName),
			},
		),
	}

	cmd.Flags().String("subject", "", SubjectUsage)
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *configCommand) describe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	return describeSchemaConfig(cmd, srClient, ctx)
}

func describeSchemaConfig(cmd *cobra.Command, srClient *srsdk.APIClient, ctx context.Context) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	var config srsdk.Config
	var resp *http.Response
	if subject != "" {
		config, resp, err = srClient.DefaultApi.GetSubjectLevelConfig(ctx, subject, nil)
		if err != nil {
			return errors.CatchNoSubjectLevelConfigError(err, resp, subject)
		}
	} else {
		config, _, err = srClient.DefaultApi.GetTopLevelConfig(ctx)
		if err != nil {
			return err
		}
	}

	outputWriter, err := output.NewListOutputWriter(cmd, []string{"CompatibilityLevel"}, []string{"Compatibility Level"}, []string{"compatibility_level"})
	if err != nil {
		return err
	}
	outputWriter.AddElement(&config)
	return outputWriter.Out()
}
