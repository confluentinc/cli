package schemaregistry

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type kekHumanOut struct {
	Name     string `human:"Name"`
	KmsType  string `human:"KMS Type"`
	KmsKeyId string `human:"KMS Key ID"`
	KmsProps string `human:"KMS Props"` // how to print this? Make it a []string of key:value?
	Doc      string `human:"Doc"`
	Shared   bool   `human:"Shared"`
	Ts       int64  `human:"TS"`
	Deleted  bool   `human:"Deleted"`
}

type kekSerializedOut struct {
	Name     string            `serialized:"name,omitempty"`
	KmsType  string            `serialized:"kmsType,omitempty"`
	KmsKeyId string            `serialized:"kmsKeyId,omitempty"`
	KmsProps map[string]string `serialized:"kmsProps,omitempty"` // how to print this? Make it a []string of key:value?
	Doc      string            `serialized:"doc,omitempty"`
	Shared   bool              `serialized:"shared,omitempty"`
	Ts       int64             `serialized:"ts,omitempty"`
	Deleted  bool              `serialized:"deleted,omitempty"`
}

func (c *command) newKekDescribeCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Kek.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.kekDescribe,
	}

	// all descriptions need to be updated. @RobertY
	cmd.Flags().String("name", "", "Name of the KEK.")
	cmd.Flags().Bool("deleted", false, "Include deleted KEK.")

	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) kekDescribe(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	deleted, err := cmd.Flags().GetBool("deleted")
	if err != nil {
		return err
	}

	res, err := client.DescribeKek(args[0], deleted)
	if err != nil {
		return err
	}

	return printKek(cmd, res)
}

func printKek(cmd *cobra.Command, res srsdk.Kek) error {
	table := output.NewTable(cmd)
	if output.GetFormat(cmd) == output.Human {
		var kmsPropsSlices []string
		for key, value := range res.GetKmsProps() {
			kmsPropsSlices = append(kmsPropsSlices, fmt.Sprintf("%s:%s", key, value))
		}
		table.Add(&kekHumanOut{
			Name:     res.GetName(),
			KmsType:  res.GetKmsType(),
			KmsKeyId: res.GetKmsKeyId(),
			KmsProps: strings.Join(kmsPropsSlices, ", "),
			Doc:      res.GetDoc(),
			Shared:   res.GetShared(),
			Ts:       res.GetTs(),
			Deleted:  res.GetDeleted(),
		})
	} else {
		table.Add(&kekSerializedOut{
			Name:     res.GetName(),
			KmsType:  res.GetKmsType(),
			KmsKeyId: res.GetKmsKeyId(),
			KmsProps: res.GetKmsProps(),
			Doc:      res.GetDoc(),
			Shared:   res.GetShared(),
			Ts:       res.GetTs(),
			Deleted:  res.GetDeleted(),
		})
	}
	return table.Print()
}
