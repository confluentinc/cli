package apikey

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/featureflags"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

var resourceKindToType = map[string]string{
	"Cluster":        "kafka",
	"ksqlDB":         "ksql",
	"SchemaRegistry": "schema-registry",
	"Cloud":          "cloud",
}

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
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	c.setKeyStoreIfNil()

	resourceType, clusterId, currentKey, err := c.resolveResourceId(cmd, c.V2Client)
	if err != nil {
		return err
	}
	if resourceType == resource.Cloud {
		clusterId = resource.Cloud
	}

	serviceAccount, err := cmd.Flags().GetString("service-account")
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

	if serviceAccount != "" {
		if resource.LookupType(serviceAccount) != resource.ServiceAccount {
			return errors.New(errors.BadServiceAccountIDErrorMsg)
		}
		_, ok := resourceIdToUserIdMap[serviceAccount]
		if !ok {
			return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.ServiceAccountNotFoundErrorMsg, serviceAccount), errors.ServiceAccountNotFoundSuggestions)
		}
	}

	currentUser, err := cmd.Flags().GetBool("current-user")
	if err != nil {
		return err
	}
	if currentUser {
		if serviceAccount != "" {
			return errors.Errorf(errors.ProhibitedFlagCombinationErrorMsg, "service-account", "current-user")
		}
		serviceAccount, err = c.getCurrentUserId()
		if err != nil {
			return err
		}
	}

	apiKeys, err := c.V2Client.ListApiKeys(serviceAccount, clusterId)
	if err != nil {
		return err
	}

	serviceAccountsMap := getServiceAccountsMap(serviceAccounts)
	usersMap := getUsersMap(allUsers)

	list := output.NewList(cmd)
	for _, apiKey := range apiKeys {
		// ignore keys owned by Confluent-internal user (healthcheck, etc)
		if !apiKey.Spec.HasOwner() {
			continue
		}

		ownerId := apiKey.Spec.Owner.GetId()
		email := c.getEmail(ownerId, resourceIdToUserIdMap, usersMap, serviceAccountsMap)

		resources := []apikeysv2.ObjectReference{apiKey.Spec.GetResource()}

		// Check if multicluster keys are enabled, and if so check the resources field
		if featureflags.Manager.BoolVariation("cli.multicluster-api-keys.enable", c.Context, v1.CliLaunchDarklyClient, true, false) {
			resources = apiKey.Spec.GetResources()
		}

		// Note that if more resource types are added with no logical clusters, then additional logic
		// needs to be added here to determine the resource type.
		for _, res := range resources {
			list.Add(&out{
				IsCurrent:    clusterId != "" && apiKey.GetId() == currentKey,
				Key:          apiKey.GetId(),
				Description:  apiKey.Spec.GetDescription(),
				OwnerId:      ownerId,
				OwnerEmail:   email,
				ResourceType: resourceKindToType[res.GetKind()],
				ResourceId:   getApiKeyResourceId(res.GetId()),
				Created:      apiKey.Metadata.GetCreatedAt().Format(time.RFC3339),
			})
		}
	}
	return list.Print()
}

func getServiceAccountsMap(serviceAccounts []iamv2.IamV2ServiceAccount) map[string]bool {
	saMap := make(map[string]bool)
	for _, sa := range serviceAccounts {
		saMap[*sa.Id] = true
	}
	return saMap
}

func getUsersMap(users []*ccloudv1.User) map[int32]*ccloudv1.User {
	userMap := make(map[int32]*ccloudv1.User)
	for _, user := range users {
		userMap[user.Id] = user
	}
	return userMap
}

func mapResourceIdToUserId(users []*ccloudv1.User) map[string]int32 {
	idMap := make(map[string]int32)
	for _, user := range users {
		idMap[user.ResourceId] = user.Id
	}
	return idMap
}

func (c *command) getEmail(resourceId string, resourceIdToUserIdMap map[string]int32, usersMap map[int32]*ccloudv1.User, serviceAccountsMap map[string]bool) string {
	if _, ok := serviceAccountsMap[resourceId]; ok {
		return "<service account>"
	}

	userId := resourceIdToUserIdMap[resourceId]
	if auditLog := v1.GetAuditLog(c.Context.Context); auditLog != nil && auditLog.GetServiceAccountId() == userId {
		return "<auditlog service account>"
	}

	if user, ok := usersMap[userId]; ok {
		return user.Email
	}

	// Check if API key is owned by the current user
	if user := c.State.Auth.User; user.ResourceId == resourceId {
		return user.GetEmail()
	}

	return "<deactivated user>"
}

func getApiKeyResourceId(id string) string {
	if id == "cloud" {
		return ""
	}
	return id
}
