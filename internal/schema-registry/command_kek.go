package schemaregistry

import (
	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/properties"
)

const (
	kmsPropsFormatSuggestions = "`--kms-properties` must be formatted as \"<key>=<value>\"."
)

func (c *command) newKekCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "kek",
		Short:       "Manage Schema Registry Key Encryption Keys (KEKs).",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(c.newKekCreateCommand(cfg))
	cmd.AddCommand(c.newKekDeleteCommand(cfg))
	cmd.AddCommand(c.newKekDescribeCommand(cfg))
	cmd.AddCommand(c.newKekListCommand(cfg))
	cmd.AddCommand(c.newKekUndeleteCommand(cfg))
	cmd.AddCommand(c.newKekUpdateCommand(cfg))

	return cmd
}

type kekOut struct {
	Name          string `human:"Name,omitempty" serialized:"name,omitempty" `
	KmsType       string `human:"KMS Type,omitempty" serialized:"kms_type,omitempty"`
	KmsKeyId      string `human:"KMS Key ID,omitempty" serialized:"kms_key_id,omitempty"`
	KmsProperties string `human:"KMS Properties,omitempty" serialized:"kms_properties,omitempty"`
	Doc           string `human:"Doc,omitempty" serialized:"doc,omitempty"`
	IsShared      bool   `human:"Shared,omitempty" serialized:"is_shared,omitempty"`
	Timestamp     int64  `human:"Timestamp,omitempty" serialized:"timestamp,omitempty"`
	IsDeleted     bool   `human:"Deleted,omitempty" serialized:"is_deleted,omitempty"`
}

func printKek(cmd *cobra.Command, kek srsdk.Kek) error {
	table := output.NewTable(cmd)
	table.Add(&kekOut{
		Name:          kek.GetName(),
		KmsType:       kek.GetKmsType(),
		KmsKeyId:      kek.GetKmsKeyId(),
		KmsProperties: convertMapToString(kek.GetKmsProps()),
		Doc:           kek.GetDoc(),
		IsShared:      kek.GetShared(),
		Timestamp:     kek.GetTs(),
		IsDeleted:     kek.GetDeleted(),
	})
	return table.Print()
}

func constructKmsProps(cmd *cobra.Command) (map[string]string, error) {
	kmsProperties, err := cmd.Flags().GetStringSlice("kms-properties")
	if err != nil {
		return nil, err
	}

	kmsPropertiesMap, err := properties.ConfigSliceToMap(kmsProperties)
	if err != nil {
		return nil, errors.NewErrorWithSuggestions(err.Error(), kmsPropsFormatSuggestions)
	}
	return kmsPropertiesMap, nil
}
