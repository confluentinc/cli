package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	serviceAccountListFields           = []string{"ResourceId", "ServiceName", "ServiceDescription"}
	serviceAccountListHumanLabels      = []string{"ID", "Name", "Description"}
	serviceAccountListStructuredLabels = []string{"id", "name", "description"}
)

func (c *serviceAccountCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service accounts.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *serviceAccountCommand) list(cmd *cobra.Command, _ []string) error {
	resp, _, err := c.V2Client.ListIamServiceAccounts()
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, serviceAccountListFields, serviceAccountListHumanLabels, serviceAccountListStructuredLabels)
	if err != nil {
		return err
	}
	for _, u := range resp.Data {
		element := &serviceAccountStruct{ResourceId: *u.Id, ServiceName: *u.DisplayName, ServiceDescription: *u.Description}
		outputWriter.AddElement(element)
	}
	return outputWriter.Out()
}
