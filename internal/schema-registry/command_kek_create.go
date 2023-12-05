package schemaregistry

import (
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
)

func (c *command) newKekCreateCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Kek.",
		Args:  cobra.NoArgs,
		RunE:  c.kekCreate,
	}

	// all descriptions need to be updated. @RobertY
	// what are required?
	cmd.Flags().String("name", "", "Name of the KEK.")
	cmd.Flags().String("kms-type", "", "The type of KMS.")
	cmd.Flags().String("kms-key-id", "", "The key ID of KMS.")
	cmd.Flags().StringSlice("kms-props", nil, "A comma-separated list?")
	cmd.Flags().String("doc", "", "")
	cmd.Flags().Bool("shared", false, "")

	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd) // guess it's needed?
	}
	pcmd.AddOutputFlag(cmd)

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

	kmsId, err := cmd.Flags().GetString("kms-key-id")
	if err != nil {
		return err
	}

	kmsPropsSlices, err := cmd.Flags().GetStringSlice("kms-props")
	if err != nil {
		return err
	}

	// construct the map
	kmsProps := make(map[string]string)
	for _, item := range kmsPropsSlices {
		pair := strings.Split(item, ":")
		if len(pair) != 2 {
			return errors.NewErrorWithSuggestions(kmsPropsFormatErrorMsg, kmsPropsFormatSuggestions)
		}
		kmsProps[pair[0]] = pair[1]
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
