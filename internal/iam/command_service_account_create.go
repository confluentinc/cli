package iam

import (
	"github.com/spf13/cobra"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
)

const (
	nameLength        = 64
	descriptionLength = 128
)

func (c *serviceAccountCommand) newCreateCommand(cfg *config.Config) *cobra.Command {
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
	isResourceOwnerEnabled := cfg.IsTest ||
		(cfg.Context() != nil &&
			featureflags.Manager.BoolVariation("auth.workload_identity.resource_owner.enabled", cfg.Context(), featureflags.GetCcloudLaunchDarklyClient(cfg.Context().PlatformName), true, false))
	cmd.Flags().String("description", "", "Description of the service account.")
	if isResourceOwnerEnabled {
		addResourceOwnerFlag(cmd, c.AuthenticatedCLICommand)
	}
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

	ldClient := featureflags.GetCcloudLaunchDarklyClient(c.Context.PlatformName)
	isResourceOwnerEnabled := c.Config.IsTest ||
		featureflags.Manager.BoolVariation("auth.workload_identity.resource_owner.enabled", c.Context, ldClient, true, false)
	resourceOwner, err := "", nil
	if isResourceOwnerEnabled {
		resourceOwner, err = cmd.Flags().GetString("resource-owner")
		if err != nil {
			return err
		}
	}
	if err := requireLen(description, descriptionLength, "description"); err != nil {
		return err
	}

	createServiceAccount := iamv2.IamV2ServiceAccount{
		DisplayName: iamv2.PtrString(name),
		Description: iamv2.PtrString(description),
	}
	serviceAccount, httpResp, err := c.V2Client.CreateIamServiceAccount(createServiceAccount, resourceOwner)
	if err != nil {
		return errors.CatchServiceNameInUseError(err, httpResp, name)
	}

	return printServiceAccount(cmd, serviceAccount)
}
