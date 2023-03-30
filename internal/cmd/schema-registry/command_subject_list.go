package schemaregistry

import (
	"context"
	"fmt"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/version"
)

type subjectListOut struct {
	Subject string `human:"Subject" serialized:"subject"`
}

func (c *command) newSubjectListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subjects.",
		Args:  cobra.NoArgs,
		RunE:  c.subjectList,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all available subjects.",
				Code: fmt.Sprintf("%s schema-registry subject list", version.CLIName),
			},
		),
	}

	cmd.Flags().Bool("deleted", false, "View the deleted subjects.")
	cmd.Flags().String("prefix", ":*:", "Subject prefix.")
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) subjectList(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	return listSubjects(cmd, srClient, ctx)
}

func listSubjects(cmd *cobra.Command, srClient *srsdk.APIClient, ctx context.Context) error {
	deleted, err := cmd.Flags().GetBool("deleted")
	if err != nil {
		return err
	}

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	listOpts := srsdk.ListOpts{
		Deleted:       optional.NewBool(deleted),
		SubjectPrefix: optional.NewString(prefix),
	}
	subjects, _, err := srClient.DefaultApi.List(ctx, &listOpts)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, subject := range subjects {
		list.Add(&subjectListOut{Subject: subject})
	}
	return list.Print()
}
