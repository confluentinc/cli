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
		Short: "Delete a Data Encryption Key (DEK).",
		Args:  cobra.NoArgs,
		RunE:  c.dekDelete,
	}

	cmd.Flags().String("kek-name", "", "Name of the Key Encryption Key (KEK).")
	cmd.Flags().String("subject", "", "Subject of the Data Encryption Key (DEK).")
	pcmd.AddAlgorithmFlag(cmd)
	cmd.Flags().String("version", "", "Version of the Data Encryption Key (DEK). When not specified, all versions of the Data Encryption Key (DEK) will be deleted.")
	cmd.Flags().Bool("permanent", false, "Delete the Data Encryption Key (DEK) permanently.")
	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}

	cobra.CheckErr(cmd.MarkFlagRequired("kek-name"))
	cobra.CheckErr(cmd.MarkFlagRequired("subject"))

	return cmd
}

func (c *command) dekDelete(cmd *cobra.Command, _ []string) error {
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

	permanent, err := cmd.Flags().GetBool("permanent")
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf("Are you sure you want to delete the %s corresponding to these parameters?", resource.Dek)
	if err := deletion.ConfirmPromptYesOrNo(cmd, promptMsg); err != nil {
		return err
	}

	if version == "" {
		err = client.DeleteDekVersions(name, subject, algorithm, permanent)
	} else {
		err = client.DeleteDekVersion(name, subject, version, algorithm, permanent)
	}
	if err != nil {
		return err
	}

	output.ErrPrintf(c.Config.EnableColor, "Deleted the %s corresponding to the parameters.\n", resource.Dek)
	return nil
}
