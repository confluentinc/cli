package schemaregistry

import (
	"context"
	"fmt"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"
)

func (c *configCommand) newDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <subject>",
		Short: "Describe the config of a subject, or at global level.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  pcmd.NewCLIRunE(c.onPremDescribe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the config of a given-name subject.",
				Code: fmt.Sprintf("%s schema-registry config describe <subject-name> %s", pversion.CLIName, errors.OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	cmd.Flags().String("sr-endpoint", "", "The URL of the schema registry cluster.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *configCommand) onPremDescribe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetAPIClientWithToken(cmd, nil, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return describeSchemaConfig(cmd, srClient, ctx)
}

func describeSchemaConfig(cmd *cobra.Command, srClient *srsdk.APIClient, ctx context.Context) error {
	subject, err := cmd.Flags().GetString("subject")
	var config srsdk.Config
	if err != nil {
		return err
	}

	if subject != "" {
		config, _, err = srClient.DefaultApi.GetSubjectLevelConfig(ctx, subject, nil)
		if err != nil {
			return err
		}
	} else {
		config, _, err = srClient.DefaultApi.GetTopLevelConfig(ctx)
		if err != nil {
			return err
		}
	}
	level := config.CompatibilityLevel

	outputOption, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}
	if outputOption == output.Human.String() {
		printConfig(config.CompatibilityLevel)
	} else {
		structuredOutput := &struct{ CompatibilityLevel string }{level}
		fields := []string{"CompatibilityLevel"}
		structuredRenames := map[string]string{"CompatibilityLevel": "compatibilityLevel"}
		return output.DescribeObject(cmd, structuredOutput, fields, map[string]string{}, structuredRenames)
	}
	return nil
}
