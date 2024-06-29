package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newSchemaDeleteCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete one or more schema versions.",
		Long:  "Delete one or more schema versions. This command should only be used if absolutely necessary.",
		Args:  cobra.NoArgs,
		RunE:  c.schemaDelete,
	}

	example := examples.Example{
		Text: `Soft delete the latest version of subject "payments".`,
		Code: "confluent schema-registry schema delete --subject payments --version latest",
	}
	if cfg.IsOnPremLogin() {
		example.Code += " " + onPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(example)

	cmd.Flags().String("subject", "", subjectUsage)
	cmd.Flags().String("version", "", `Version of the schema. Can be a specific version, "all", or "latest".`)
	cmd.Flags().Bool("permanent", false, "Permanently delete the schema.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}
	pcmd.AddForceFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("subject"))
	cobra.CheckErr(cmd.MarkFlagRequired("version"))

	return cmd
}

func (c *command) schemaDelete(cmd *cobra.Command, _ []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
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

	permanent, err := cmd.Flags().GetBool("permanent")
	if err != nil {
		return err
	}

	checkVersion := version
	if version == "all" {
		// check that at least one version for the input subject exists
		checkVersion = "latest"
	}
	if permanent {
		if checkVersion != "latest" {
			if _, err := client.GetSchemaByVersion(subject, checkVersion, true); err != nil {
				return catchSchemaNotFoundError(err, subject, checkVersion)
			} else if _, err := client.GetSchemaByVersion(subject, checkVersion, false); err == nil {
				return fmt.Errorf("you must first soft delete a schema version before you can permanently delete it")
			}
		}
	} else if _, err := client.GetSchemaByVersion(subject, checkVersion, false); err != nil {
		return catchSchemaNotFoundError(err, subject, checkVersion)
	}

	subjectWithVersion := fmt.Sprintf("%s (version %s)", subject, version)
	promptMsg := fmt.Sprintf("Are you sure you want to delete schema \"%s\"?", subjectWithVersion)
	if permanent {
		promptMsg = fmt.Sprintf("Are you sure you want to permanently delete schema \"%s\"?", subjectWithVersion)
	}
	if err := deletion.ConfirmPrompt(cmd, promptMsg); err != nil {
		return err
	}

	deleteType := "soft"
	if permanent {
		deleteType = "hard"
	}

	var versions []int32
	if version == "all" {
		v, err := client.DeleteSubject(subject, permanent)
		if err != nil {
			return catchSchemaNotFoundError(err, subject, version)
		}
		output.Printf(c.Config.EnableColor, "Successfully %s deleted all versions for subject \"%s\".\n", deleteType, subject)
		versions = v
	} else {
		v, err := client.DeleteSchemaVersion(subject, version, permanent)
		if err != nil {
			return catchSchemaNotFoundError(err, subject, version)
		}
		output.Printf(c.Config.EnableColor, "Successfully %s deleted version \"%s\" for subject \"%s\".\n", deleteType, version, subject)
		versions = []int32{v}
	}

	list := output.NewList(cmd)
	for _, version := range versions {
		list.Add(&versionOut{Version: version})
	}
	return list.Print()
}
