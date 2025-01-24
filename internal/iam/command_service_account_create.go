package iam

import (
	"fmt"
	"github.com/confluentinc/cli/v4/pkg/utils"
	"github.com/spf13/cobra"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

const (
	nameLength        = 64
	descriptionLength = 128
)

func (c *serviceAccountCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a service account.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a service account named "my-service-account".`,
				Code: `confluent iam service-account create my-service-account --description "new description"`,
			},
		),
	}

	cmd.Flags().String("description", "", "Description of the service account.")
	items := []string{"user", "group-mapping", "service-account", "identity-pool"}
	cmd.Flags().String("resource-owner", "", fmt.Sprintf("The resource_id of the principal who will be assigned resource owner on the "+
		"created service account. Principal can be a %s.",
		utils.ArrayToCommaDelimitedString(items, "or")))
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("description"))

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

	assignedResourceOwner, err := cmd.Flags().GetString("resource-owner")
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
	serviceAccount, httpResp, err := c.V2Client.CreateIamServiceAccount(createServiceAccount, assignedResourceOwner)
	if err != nil {
		return errors.CatchServiceNameInUseError(err, httpResp, name)
	}

	return printServiceAccount(cmd, serviceAccount)
}
