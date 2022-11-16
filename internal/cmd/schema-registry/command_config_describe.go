package schemaregistry

import (
	"context"
	"fmt"
	"net/http"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

type configOut struct {
	CompatibilityLevel string `human:"Compatibility Level" serialized:"compatibility_level"`
}

func (c *configCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe top-level or subject-level schema compatibility.",
		Args:  cobra.NoArgs,
		RunE:  c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the configuration of subject "payments".`,
				Code: fmt.Sprintf("%s schema-registry config describe --subject payments", pversion.CLIName),
			},
			examples.Example{
				Text: "Describe the top-level configuration.",
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
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
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
	var httpResp *http.Response
	if subject != "" {
		config, httpResp, err = srClient.DefaultApi.GetSubjectLevelConfig(ctx, subject, nil)
		if err != nil {
			return errors.CatchNoSubjectLevelConfigError(err, httpResp, subject)
		}
	} else {
		config, _, err = srClient.DefaultApi.GetTopLevelConfig(ctx)
		if err != nil {
			return err
		}
	}

	table := output.NewTable(cmd)
	table.Add(&configOut{CompatibilityLevel: config.CompatibilityLevel})
	return table.Print()
}
