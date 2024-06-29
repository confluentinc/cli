package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

func (c *command) newDekDescribeCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe a Data Encryption Key (DEK).",
		Args:  cobra.NoArgs,
		RunE:  c.dekDescribe,
	}

	cmd.Flags().String("kek-name", "", "Name of the Key Encryption Key (KEK).")
	cmd.Flags().String("subject", "", "Subject of the Data Encryption Key (DEK).")
	pcmd.AddAlgorithmFlag(cmd)
	cmd.Flags().String("version", "1", "Version of the Data Encryption Key (DEK).")
	cmd.Flags().Bool("all", false, "Include soft-deleted Data Encryption Key (DEK).")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("kek-name"))
	cobra.CheckErr(cmd.MarkFlagRequired("subject"))

	return cmd
}

func (c *command) dekDescribe(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("kek-name")
	if err != nil {
		return err
	}

	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	algorithm, err := cmd.Flags().GetString("algorithm")
	if err != nil {
		return err
	}

	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}

	dek, err := client.GetDekByVersion(name, subject, version, algorithm, all)
	if err != nil {
		return err
	}

	return printDek(cmd, dek)
}
