package apikey

import (
	"fmt"
	"time"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/spf13/cobra"
)

var (
	listFields           = []string{"Key", "Description", "UserResourceId", "UserEmail", "ResourceType", "ResourceId", "Created"}
	listHumanLabels      = []string{"Key", "Description", "Owner Resource ID", "Owner Email", "Resource Type", "Resource ID", "Created"}
	listStructuredLabels = []string{"key", "description", "owner_resource_id", "owner_email", "resource_type", "resource_id", "created"}
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the API keys.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List the API keys that belong to service account "sa-123456" on cluster "lkc-123456".`,
				Code: "confluent api-key list --resource lkc-123456 --service-account sa-123456",
			},
		),
	}

	cmd.Flags().String(resourceFlagName, "", `The resource ID to filter by. Use "cloud" to show only Cloud API keys.`)
	cmd.Flags().Bool("current-user", false, "Show only API keys belonging to current user.")
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	c.setKeyStoreIfNil()

	resourceType, clusterId, currentKey, err := c.resolveResourceId(cmd, c.Client)
	if err != nil {
		return err
	}
	if resourceType == resource.Cloud {
		clusterId = resource.Cloud
	}

	ownerResourceId, err := cmd.Flags().GetString("service-account")
	if err != nil {
		return err
	}

	serviceAccounts, err := c.V2Client.ListIamServiceAccounts()
	if err != nil {
		return err
	}
	allUsers, err := c.getAllUsers()
	if err != nil {
		return err
	}
	resourceIdToUserIdMap := mapResourceIdToUserId(allUsers)

	if ownerResourceId != "" {
		if resource.LookupType(ownerResourceId) != resource.ServiceAccount {
			return errors.New(errors.BadServiceAccountIDErrorMsg)
		}
		_, ok := resourceIdToUserIdMap[ownerResourceId]
		if !ok {
			return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.ServiceAccountNotFoundErrorMsg, ownerResourceId), errors.ServiceAccountNotFoundSuggestions)
		}
	}

	currentUser, err := cmd.Flags().GetBool("current-user")
	if err != nil {
		return err
	}
	if currentUser {
		if ownerResourceId != "" {
			return errors.Errorf(errors.ProhibitedFlagCombinationErrorMsg, "service-account", "current-user")
		}
		ownerResourceId, err = c.getCurrentUserId()
		if err != nil {
			return err
		}
	}

	apiKeys, err := c.V2Client.ListApiKeys(ownerResourceId, clusterId)
	if err != nil {
		return err
	}

	serviceAccountsMap := getServiceAccountsMap(serviceAccounts)
	usersMap := getUsersMap(allUsers)

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}

	for _, apiKey := range apiKeys {
		// ignore keys owned by Confluent-internal user (healthcheck, etc)
		if !apiKey.Spec.HasOwner() {
			continue
		}

		// Add '*' only in the case where we are printing out tables
		outputKey := *apiKey.Id
		if outputWriter.GetOutputFormat() == output.Human {
			if clusterId != "" && *apiKey.Id == currentKey {
				outputKey = fmt.Sprintf("* %s", *apiKey.Id)
			} else {
				outputKey = fmt.Sprintf("  %s", *apiKey.Id)
			}
		}

		ownerId := apiKey.GetSpec().Owner.GetId()
		email := c.getEmail(ownerId, resourceIdToUserIdMap, usersMap, serviceAccountsMap)

		// Note that if more resource types are added with no logical clusters, then additional logic
		// needs to be added here to determine the resource type.
		outputWriter.AddElement(&apiKeyRow{
			Key:            outputKey,
			Description:    *apiKey.GetSpec().Description,
			UserResourceId: ownerId,
			UserEmail:      email,
			ResourceType:   resourceKindToType[apiKey.GetSpec().Resource.GetKind()],
			ResourceId:     getApiKeyResourceId(apiKey),
			Created:        apiKey.GetMetadata().CreatedAt.Format(time.RFC3339),
		})
	}

	return outputWriter.Out()
}

func getServiceAccountsMap(serviceAccounts []iamv2.IamV2ServiceAccount) map[string]bool {
	saMap := make(map[string]bool)
	for _, sa := range serviceAccounts {
		saMap[*sa.Id] = true
	}
	return saMap
}

func getUsersMap(users []*orgv1.User) map[int32]*orgv1.User {
	userMap := make(map[int32]*orgv1.User)
	for _, user := range users {
		userMap[user.Id] = user
	}
	return userMap
}

func mapResourceIdToUserId(users []*orgv1.User) map[string]int32 {
	idMap := make(map[string]int32)
	for _, user := range users {
		idMap[user.ResourceId] = user.Id
	}
	return idMap
}

func (c *command) getEmail(resourceId string, resourceIdToUserIdMap map[string]int32, usersMap map[int32]*orgv1.User, serviceAccountsMap map[string]bool) string {
	if _, ok := serviceAccountsMap[resourceId]; ok {
		return "<service account>"
	}

	userId := resourceIdToUserIdMap[resourceId]
	if auditLog, ok := pcmd.AreAuditLogsEnabled(c.State); ok && auditLog.ServiceAccountId == userId {
		return "<auditlog service account>"
	}

	if user, ok := usersMap[userId]; ok {
		return user.Email
	}

	return "<deactivated user>"
}

func getApiKeyResourceId(apiKey apikeysv2.IamV2ApiKey) string {
	id := apiKey.GetSpec().Resource.GetId()
	if id == "cloud" {
		return ""
	}
	return id
}
