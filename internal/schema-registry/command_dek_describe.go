package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

func (c *command) newDekDescribeCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe a Dek.",
		Args:  cobra.NoArgs,
		RunE:  c.dekDescribe,
	}

	cmd.Flags().String("name", "", "Name of the KEK.")
	cmd.Flags().String("subject", "", "Subject of the DEK.")
	pcmd.AddAlgorithmFlag(cmd)
	cmd.Flags().String("version", "1", "Version of the DEK.")
	cmd.Flags().Bool("deleted", false, "Include deleted DEK.")

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

	return cmd
}

func (c *command) dekDescribe(cmd *cobra.Command, args []string) error {
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

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	algorithm, err := cmd.Flags().GetString("algorithm")
	if err != nil {
		return err
	}

	deleted, err := cmd.Flags().GetBool("deleted")
	if err != nil {
		return err
	}

	dek, err := client.GetDekByVersion(name, subject, version, algorithm, deleted)
	if err != nil {
		return err
	}

	return printDek(cmd, dek)
}
