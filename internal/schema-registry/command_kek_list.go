package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newKekListCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Key Encryption Keys (KEKs).",
		Args:  cobra.NoArgs,
		RunE:  c.kekList,
	}

	cmd.Flags().Bool("all", false, "Include soft-deleted Key Encryption Keys (KEKs).")
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

func (c *command) kekList(cmd *cobra.Command, _ []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}

	keks, err := client.ListKeks(all)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, kekName := range keks {
		kek, err := client.DescribeKek(kekName, all)
		if err != nil {
			return err
		}
		list.Add(&kekOut{
			Name:          kekName,
			KmsType:       kek.GetKmsType(),
			KmsKeyId:      kek.GetKmsKeyId(),
			KmsProperties: convertMapToString(kek.GetKmsProps()),
			Doc:           kek.GetDoc(),
			IsShared:      kek.GetShared(),
			Timestamp:     kek.GetTs(),
			IsDeleted:     kek.GetDeleted(),
		})
	}
	return list.Print()
}
