package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newDekSubjectListCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Schema Registry Data Encryption Key (DEK) subjects.",
		Args:  cobra.NoArgs,
		RunE:  c.dekSubjectList,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List subjects for the Data Encryption Key (DEK) created with a KEK named "test":`,
				Code: "confluent schema-registry dek subject list --name test",
			},
		),
	}

	cmd.Flags().String("kek-name", "", "Name of the Key Encryption Key (KEK).")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("kek-name"))

	return cmd
}
func (c *command) dekSubjectList(cmd *cobra.Command, _ []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("kek-name")
	if err != nil {
		return err
	}

	subjects, err := client.GetDekSubjects(name)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, subject := range subjects {
		list.Add(&subjectListOut{Subject: subject})
	}
	return list.Print()
}
