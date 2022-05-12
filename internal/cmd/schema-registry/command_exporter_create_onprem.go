package schemaregistry

import (
	"fmt"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	"github.com/spf13/cobra"
)

func (c *exporterCommand) newCreateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "create <name>",
		Short:       "Create a new schema exporter.",
		Args:        cobra.ExactArgs(1),
		RunE:        c.onPremCreate,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a new schema exporter.",
				Code: fmt.Sprintf(`%s schema-registry exporter create my-exporter --config-file config.txt --subjects my-subject1,my-subject2 --subject-format my-\${subject} --context-type CUSTOM --context-name my-context %s`, pversion.CLIName, OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().String("config-file", "", "Exporter config file.")
	cmd.Flags().StringSlice("subjects", []string{"*"}, "Exporter subjects. Use a comma separated list, or specify the flag multiple times.")
	cmd.Flags().String("subject-format", "${subject}", "Exporter subject rename format. The format string can contain ${subject}, which will be replaced with default subject name.")
	addContextTypeFlag(cmd)
	cmd.Flags().String("context-name", "", "Exporter context name.")
	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("config-file")

	return cmd
}

func (c *exporterCommand) onPremCreate(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetSrApiClientWithToken(cmd, nil, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return createExporter(cmd, args[0], srClient, ctx)
}
