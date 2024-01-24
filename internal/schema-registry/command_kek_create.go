package schemaregistry

import (
	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newKekCreateCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Key Encryption Key (KEK).",
		Args:  cobra.NoArgs,
		RunE:  c.kekCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a KEK with a AWS KMS key:`,
				Code: "confluent schema-registry kek create --name test --kms-type AWS_KMS --kms-key arn:aws:kms:us-west-2:037502941121:key/a1231e22-1n78-4l0d-9d50-9pww5faedb54 --kms-properties KeyUsage=ENCRYPT_DECRYPT,KeyState=Enabled",
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the Key Encryption Key (KEK).")
	pcmd.AddKmsTypeFlag(cmd)
	cmd.Flags().String("kms-key", "", "The key ID of the Key Management Service (KMS).")
	cmd.Flags().StringSlice("kms-properties", nil, "A comma-separated list of additional properties (key=value) used to access the Key Management Service (KMS).")
	cmd.Flags().String("doc", "", "An optional user-friendly description for the Key Encryption Key (KEK).")
	cmd.Flags().Bool("shared", false, "If the DEK Registry has shared access to the Key Management Service (KMS).")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("name"))
	cobra.CheckErr(cmd.MarkFlagRequired("kms-type"))
	cobra.CheckErr(cmd.MarkFlagRequired("kms-key"))

	return cmd
}

func (c *command) kekCreate(cmd *cobra.Command, _ []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	kmsType, err := cmd.Flags().GetString("kms-type")
	if err != nil {
		return err
	}

	kmsId, err := cmd.Flags().GetString("kms-key")
	if err != nil {
		return err
	}

	kmsProps, err := constructKmsProps(cmd)
	if err != nil {
		return err
	}

	doc, err := cmd.Flags().GetString("doc")
	if err != nil {
		return err
	}

	shared, err := cmd.Flags().GetBool("shared")
	if err != nil {
		return err
	}

	createReq := srsdk.CreateKekRequest{
		Name:     srsdk.PtrString(name),
		KmsType:  srsdk.PtrString(kmsType),
		KmsKeyId: srsdk.PtrString(kmsId),
		KmsProps: &kmsProps,
		Doc:      srsdk.PtrString(doc),
		Shared:   srsdk.PtrBool(shared),
	}

	res, err := client.CreateKek(name, createReq)
	if err != nil {
		return err
	}

	return printKek(cmd, res)
}
