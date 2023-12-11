package schemaregistry

import (
	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newDekCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "dek",
		Short:       "Manage Schema Registry DEK.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(c.newDekCreateCommand(cfg))
	cmd.AddCommand(c.newDekDeleteCommand(cfg))
	cmd.AddCommand(c.newDekDescribeCommand(cfg))
	cmd.AddCommand(c.newDekSubjectCommand(cfg))
	cmd.AddCommand(c.newDekUndeleteCommand(cfg))
	cmd.AddCommand(c.newDekVersionCommand(cfg))

	return cmd
}

type dekOut struct {
	Name                 string `human:"Name" json:"name"`
	Subject              string `human:"Subject" json:"subject"`
	Version              int32  `human:"Version" json:"version"`
	Algorithm            string `human:"Algorithm" json:"algorithm"`
	EncryptedKeyMaterial string `human:"Encrypted Key Material" json:"encrypted_key_material"`
	KeyMaterial          string `human:"Key Material" json:"key_material"`
	Timestamp            int64  `human:"Timestamp" json:"timestamp"`
}

func printDek(cmd *cobra.Command, dek srsdk.Dek) error {
	table := output.NewTable(cmd)
	table.Add(&dekOut{
		Name:                 dek.GetKekName(),
		Subject:              dek.GetSubject(),
		Version:              dek.GetVersion(),
		Algorithm:            dek.GetAlgorithm(),
		EncryptedKeyMaterial: dek.GetEncryptedKeyMaterial(),
		KeyMaterial:          dek.GetKeyMaterial(),
		Timestamp:            dek.GetTs(),
	})
	return table.Print()
}
