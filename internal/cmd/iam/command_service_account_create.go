package iam

import (
	"github.com/spf13/cobra"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

const (
	nameLength        = 64
	descriptionLength = 128
)

var (
	describeFields            = []string{"ResourceId", "Name", "Description"}
	describeHumanRenames      = map[string]string{"Name": "Name", "Description": "Description", "ResourceId": "ID"}
	describeStructuredRenames = map[string]string{"Name": "name", "Description": "description", "ResourceId": "id"}
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

	createServiceAccount := iamv2.IamV2ServiceAccount{
		DisplayName: iamv2.PtrString(name),
		Description: iamv2.PtrString(description),
	}
	createResp, resp, err := c.V2Client.CreateIamServiceAccount(createServiceAccount)
	if err != nil {
		return errors.CatchServiceNameInUseError(err, resp, name)
	}

	DescribeServiceAccount := &serviceAccount{ResourceId: *createResp.Id, Name: *createResp.DisplayName, Description: *createResp.Description}

	return output.DescribeObject(cmd, DescribeServiceAccount, describeFields, describeHumanRenames, describeStructuredRenames)
}
