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
		Short: "List Schema Registry DEK versions.",
		Args:  cobra.NoArgs,
		RunE:  c.dekVersionList,
	}

	cmd.Flags().String("name", "", "Name of the KEK.")
	cmd.Flags().String("subject", "", "Subject of the DEK.")
	pcmd.AddAlgorithmFlag(cmd)
	cmd.Flags().Bool("deleted", false, "Include deleted DEK.")

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

	name, err := cmd.Flags().GetString("name")
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

	deleted, err := cmd.Flags().GetBool("deleted")
	if err != nil {
		return err
	}

	versions, err := client.GetDeKVersions(name, subject, algorithm, deleted)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, version := range versions {
		list.Add(&versionOut{Version: version})
	}
	return list.Print()
}
