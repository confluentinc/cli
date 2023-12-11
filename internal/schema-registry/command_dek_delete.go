package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newDekDeleteCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a DEK.",
		Args:  cobra.NoArgs,
		RunE:  c.dekDelete,
	}

	cmd.Flags().String("name", "", "Name of the KEK.")
	cmd.Flags().String("subject", "", "Subject of the DEK.")
	pcmd.AddAlgorithmFlag(cmd)
	cmd.Flags().String("version", "", "Version of the DEK. When not specified, all verions of DEK will be deleted.")
	cmd.Flags().Bool("permanent", false, "Delete DEK permanently.")
	pcmd.AddForceFlag(cmd)

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

func (c *command) dekDelete(cmd *cobra.Command, _ []string) error {
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

	permanent, err := cmd.Flags().GetBool("permanent")
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf("Are you sure you want to delete the %s corresponding to these parameters?", resource.Dek)
	if err := deletion.ConfirmDeletionYesNo(cmd, promptMsg); err != nil {
		return err
	}

	var deleteErr error
	if version == "" {
		deleteErr = client.DeleteDekVersions(name, subject, algorithm, permanent)
	} else {
		deleteErr = client.DeleteDekVersion(name, subject, version, algorithm, permanent)
	}
	if deleteErr != nil {
		return deleteErr
	}

	output.ErrPrintln(c.Config.EnableColor, fmt.Sprintf("Deleted the %s corresponding to the parameters.", resource.Dek))
	return nil
}
