package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type subjectListOut struct {
	Subject string `human:"Subject" serialized:"subject"`
}

func (c *command) newSubjectListCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subjects.",
		Args:  cobra.NoArgs,
		RunE:  c.subjectList,
	}

	cmd.Flags().Bool("all", false, "Include deleted subjects.")
	cmd.Flags().String("prefix", ":*:", "Subject prefix.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
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
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	subjects, err := client.List(prefix, all)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, subject := range subjects {
		list.Add(&subjectListOut{Subject: subject})
	}
	return list.Print()
}
