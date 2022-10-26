package apikey

import (
	"time"

	"github.com/spf13/cobra"

	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/featureflags"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe an API key.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()

	apiKey, r, err := c.V2Client.GetApiKey(args[0])
	if err != nil {
		return errors.CatchApiKeyForbiddenAccessError(err, getOperation, r)
	}

	var ownerId string
	var email string

	if apiKey.Spec.HasOwner() {
		allUsers, err := c.getAllUsers()
		if err != nil {
			return err
		}
		resourceIdToUserIdMap := mapResourceIdToUserId(allUsers)
		usersMap := getUsersMap(allUsers)

		serviceAccounts, err := c.V2Client.ListIamServiceAccounts()
		if err != nil {
			return err
		}
		serviceAccountsMap := getServiceAccountsMap(serviceAccounts)

		ownerId = apiKey.Spec.Owner.GetId()
		email = c.getEmail(ownerId, resourceIdToUserIdMap, usersMap, serviceAccountsMap)
	}

	resources := []apikeysv2.ObjectReference{apiKey.Spec.GetResource()}

	// Check if multicluster keys are enabled, and if so check the resources field
	if featureflags.Manager.BoolVariation("cli.multicluster-api-keys.enable", c.Context, v1.CliLaunchDarklyClient, true, false) {
		resources = apiKey.Spec.GetResources()
	}

	outputWriter, err := output.NewListOutputWriter(cmd, fields, humanLabels, structuredLabels)
	if err != nil {
		return err
	}

	// Note that if more resource types are added with no logical clusters, then additional logic
	// needs to be added here to determine the resource type.
	for _, res := range resources {
		outputWriter.AddElement(&row{
			Key:             apiKey.GetId(),
			Description:     apiKey.Spec.GetDescription(),
			OwnerResourceId: ownerId,
			OwnerEmail:      email,
			ResourceType:    resourceKindToType[res.GetKind()],
			ResourceId:      getApiKeyResourceId(res.GetId()),
			Created:         apiKey.Metadata.GetCreatedAt().Format(time.RFC3339),
		})
	}

	return outputWriter.Out()
}
