package iam

import (
	"context"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

const (
	nameLength        = 64
	descriptionLength = 128
)

var (
	describeFields            = []string{"ResourceId", "ServiceName", "ServiceDescription"}
	describeHumanRenames      = map[string]string{"ServiceName": "Name", "ServiceDescription": "Description", "ResourceId": "ID"}
	describeStructuredRenames = map[string]string{"ServiceName": "name", "ServiceDescription": "description", "ResourceId": "id"}
)

func (c *serviceAccountCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a service account.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a service account named `DemoServiceAccount`.",
				Code: `confluent iam service-account create DemoServiceAccount --description "This is a demo service account."`,
			},
		),
	}

	cmd.Flags().String("description", "", "Description of the service account.")
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("description")

	return cmd
}

func (c *serviceAccountCommand) create(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := requireLen(name, nameLength, "service name"); err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	if err := requireLen(description, descriptionLength, "description"); err != nil {
		return err
	}

	user := &orgv1.User{
		ServiceName:        name,
		ServiceDescription: description,
		ServiceAccount:     true,
	}
	user, err = c.Client.User.CreateServiceAccount(context.Background(), user)
	if err != nil {
		return err
	}
	return output.DescribeObject(cmd, user, describeFields, describeHumanRenames, describeStructuredRenames)
}
