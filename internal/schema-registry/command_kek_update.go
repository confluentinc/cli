package schemaregistry

import (
	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

func (c *command) newKekUpdateCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a KEK.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.kekUpdate,
	}

	cmd.Flags().StringSlice("kms-properties", nil, "A comma-separated list of additional properties (key=value) used to access the KMS.")
	cmd.Flags().String("doc", "", "An optional user-friendly description for the KEK.")
	cmd.Flags().Bool("shared", false, "If the DEK Registry has shared access to the KMS.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsOneRequired("kms-properties", "doc", "shared")

	return cmd
}

func (c *command) kekUpdate(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	kek, err := client.DescribeKek(args[0], false)
	if err != nil {
		return err
	}

	updateReq := srsdk.UpdateKekRequest{
		KmsProps: kek.KmsProps,
		Doc:      kek.Doc,
		Shared:   kek.Shared,
	}

	if cmd.Flags().Changed("kms-properties") {
		kmsProps, err := constructKmsProps(cmd)
		if err != nil {
			return err
		}
		updateReq.SetKmsProps(kmsProps)
	}

	if cmd.Flags().Changed("doc") {
		doc, err := cmd.Flags().GetString("doc")
		if err != nil {
			return err
		}
		updateReq.SetDoc(doc)
	}

	if cmd.Flags().Changed("shared") {
		shared, err := cmd.Flags().GetBool("shared")
		if err != nil {
			return err
		}
		updateReq.SetShared(shared)
	}

	res, err := client.UpdateKek(args[0], updateReq)
	if err != nil {
		return err
	}

	return printKek(cmd, res)
}
