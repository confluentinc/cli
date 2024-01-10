package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newDekVersionListCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Schema Registry Data Encryption Key (DEK) versions.",
		Args:  cobra.NoArgs,
		RunE:  c.dekVersionList,
	}

	cmd.Flags().String("kek-name", "", "Name of the Key Encryption Key (KEK).")
	cmd.Flags().String("subject", "", "Subject of the Data Encryption Key (DEK).")
	pcmd.AddAlgorithmFlag(cmd)
	cmd.Flags().Bool("all", false, "Include soft-deleted Data Encryption Key (DEK).")
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
func (c *command) dekVersionList(cmd *cobra.Command, args []string) error {
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

	algorithm, err := cmd.Flags().GetString("algorithm")
	if err != nil {
		return err
	}

	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}

	versions, err := client.GetDeKVersions(name, subject, algorithm, all)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, version := range versions {
		list.Add(&versionOut{Version: version})
	}
	return list.Print()
}
