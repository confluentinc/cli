package apikey

import (
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type out struct {
	IsCurrent    bool   `human:"Current,omitempty" serialized:"is_current,omitempty"`
	Key          string `human:"Key" serialized:"key"`
	Description  string `human:"Description" serialized:"description"`
	Owner        string `human:"Owner" serialized:"owner"`
	OwnerEmail   string `human:"Owner Email" serialized:"owner_email"`
	ResourceType string `human:"Resource Type" serialized:"resource_type"`
	Resource     string `human:"Resource" serialized:"resource"`
	Created      string `human:"Created" serialized:"created"`
	Expiration   string `human:"Expiration" serialized:"expiration"`
}

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe an API key.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()

	apiKey, httpResp, err := c.V2Client.GetApiKey(args[0])
	if err != nil {
		return errors.CatchApiKeyForbiddenAccessError(err, getOperation, httpResp)
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

		serviceAccounts, err := c.V2Client.ListIamServiceAccounts(nil)
		if err != nil {
			return err
		}
		serviceAccountsMap := getServiceAccountsMap(serviceAccounts)

		ownerId = apiKey.Spec.Owner.GetId()
		auditLogServiceAccountId := c.getAuditLogServiceAccountId()
		email = c.getEmail(ownerId, auditLogServiceAccountId, resourceIdToUserIdMap, usersMap, serviceAccountsMap)
	}

	resource := apiKey.Spec.GetResource()

	list := output.NewList(cmd)
	// Note that if more resource types are added with no logical clusters, then additional logic
	// needs to be added here to determine the resource type.
	list.Add(&out{
		Key:          apiKey.GetId(),
		Description:  apiKey.Spec.GetDescription(),
		Owner:        ownerId,
		OwnerEmail:   email,
		ResourceType: getResourceType(resource),
		Resource:     getResourceId(resource.GetId()),
		Created:      apiKey.Metadata.GetCreatedAt().Format(time.RFC3339),
		Expiration:   apiKey.Spec.GetExpiresAt(),
	})
	return list.Print()
}
