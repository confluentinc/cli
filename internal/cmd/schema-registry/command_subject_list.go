package schemaregistry

import (
	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type subjectListOut struct {
	Subject string `human:"Subject" serialized:"subject"`
}

func (c *command) newSubjectListCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subjects.",
		Args:  cobra.NoArgs,
		RunE:  c.subjectList,
	}

	cmd.Flags().Bool("deleted", false, "View the deleted subjects.")
	cmd.Flags().String("prefix", ":*:", "Subject prefix.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	}
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

func (c *command) subjectList(cmd *cobra.Command, _ []string) error {
	client, err := c.GetSchemaRegistryClient()
	if err != nil {
		return err
	}

	deleted, err := cmd.Flags().GetBool("deleted")
	if err != nil {
		return err
	}

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	opts := &srsdk.ListOpts{
		Deleted:       optional.NewBool(deleted),
		SubjectPrefix: optional.NewString(prefix),
	}
	subjects, err := client.List(opts)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, subject := range subjects {
		list.Add(&subjectListOut{Subject: subject})
	}
	return list.Print()
}
