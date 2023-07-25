package schemaregistry

import (
	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type versionOut struct {
	Version int32 `human:"Version" serialized:"version"`
}

func (c *command) newSubjectDescribeCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <subject>",
		Short: "Describe subject versions.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.subjectDescribe,
	}

	example := examples.Example{
		Text: `Retrieve all versions registered under subject "payments" and its compatibility level.`,
		Code: "confluent schema-registry subject describe payments",
	}
	if cfg.IsOnPremLogin() {
		example.Code += " " + OnPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(example)

	cmd.Flags().Bool("deleted", false, "View the deleted schema.")
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	}
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	if cfg.IsCloudLogin() {
		// Deprecated
		pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
		cobra.CheckErr(cmd.Flags().MarkHidden("api-key"))

		// Deprecated
		pcmd.AddApiSecretFlag(cmd)
		cobra.CheckErr(cmd.Flags().MarkHidden("api-secret"))
	}

	return cmd
}

func (c *command) subjectDescribe(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient()
	if err != nil {
		return err
	}

	deleted, err := cmd.Flags().GetBool("deleted")
	if err != nil {
		return err
	}

	opts := &srsdk.ListVersionsOpts{Deleted: optional.NewBool(deleted)}
	versions, err := client.ListVersions(args[0], opts)
	if err != nil {
		return catchSchemaNotFoundError(err, args[0], "")
	}

	list := output.NewList(cmd)
	for _, version := range versions {
		list.Add(&versionOut{Version: version})
	}
	return list.Print()
}
