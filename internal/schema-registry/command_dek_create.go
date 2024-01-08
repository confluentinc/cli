package schemaregistry

import (
	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newDekCreateCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Data Encryption Key (DEK).",
		Args:  cobra.NoArgs,
		RunE:  c.dekCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a DEK with KEK "test", and subject "test-value":`,
				Code: "confluent schema-registry dek create --name test --subject test-value --version 1",
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the KEK.")
	cmd.Flags().String("subject", "", "Subject of the Data Encryption Key (DEK).")
	pcmd.AddAlgorithmFlag(cmd)
	cmd.Flags().Int32("version", 0, "Version of the Data Encryption Key (DEK).")
	cmd.Flags().String("encrypted-key-material", "", "The encrypted key material for the Data Encryption Key (DEK).")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("name"))
	cobra.CheckErr(cmd.MarkFlagRequired("subject"))
	cobra.CheckErr(cmd.MarkFlagRequired("version"))

	return cmd
}

func (c *command) dekCreate(cmd *cobra.Command, _ []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	version, err := cmd.Flags().GetInt32("version")
	if err != nil {
		return err
	}

	algorithm, err := cmd.Flags().GetString("algorithm")
	if err != nil {
		return err
	}

	encryptedKeyMaterial, err := cmd.Flags().GetString("encrypted-key-material")
	if err != nil {
		return err
	}

	createReq := srsdk.CreateDekRequest{
		Subject:              srsdk.PtrString(subject),
		Version:              srsdk.PtrInt32(version),
		Algorithm:            srsdk.PtrString(algorithm),
		EncryptedKeyMaterial: srsdk.PtrString(encryptedKeyMaterial),
	}

	dek, err := client.CreateDek(name, createReq)
	if err != nil {
		return err
	}

	return printDek(cmd, dek)
}
