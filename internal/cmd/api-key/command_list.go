package apikey

import (
	"context"
	"fmt"
	"time"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
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
	type keyDisplay struct {
		Key            string
		Description    string
		UserResourceId string
		UserEmail      string
		ResourceType   string
		ResourceId     string
		Created        string
	}
	var apiKeys []*schedv1.ApiKey

	resourceType, clusterId, currentKey, err := c.resolveResourceId(cmd, c.Client)
	if err != nil {
		return err
	}

	var logicalClusters []*schedv1.ApiKey_Cluster
	if clusterId != "" {
		logicalClusters = []*schedv1.ApiKey_Cluster{{Id: clusterId, Type: resourceType}}
	}

	serviceAccountID, err := cmd.Flags().GetString("service-account")
	if err != nil {
		return err
	}

	serviceAccounts, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return err
	}
	users, err := c.Client.User.List(context.Background())
	if err != nil {
		return err
	}
	allUsers, err := c.getAllUsers()
	if err != nil {
		return err
	}

	userId := int32(0)
	serviceAccount := false
	if serviceAccountID != "" { // if user inputs resource ID, get corresponding numeric ID
		serviceAccount = true
		if resource.LookupType(serviceAccountID) != resource.ServiceAccount {
			return errors.New(errors.BadServiceAccountIDErrorMsg)
		}
		userIdMap := c.mapResourceIdToUserId(allUsers)
		var ok bool
		userId, ok = userIdMap[serviceAccountID]
		if !ok {
			return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.ServiceAccountNotFoundErrorMsg, serviceAccountID), errors.ServiceAccountNotFoundSuggestions)
		}
	}

	currentUser, err := cmd.Flags().GetBool("current-user")
	if err != nil {
		return err
	}
	if currentUser {
		if userId != 0 {
			return errors.Errorf(errors.ProhibitedFlagCombinationErrorMsg, "service-account", "current-user")
		}
		userId = c.State.Auth.User.Id
	}
	apiKeys, err = c.Client.APIKey.List(context.Background(), &schedv1.ApiKey{AccountId: c.EnvironmentId(), LogicalClusters: logicalClusters, UserId: userId, ServiceAccount: serviceAccount})
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}

	serviceAccountsMap := getServiceAccountsMap(serviceAccounts)
	usersMap := getUsersMap(users)
	resourceIdMap := c.mapUserIdToResourceId(allUsers)

	for _, apiKey := range apiKeys {
		// ignore keys owned by Confluent-internal user (healthcheck, etc)
		if apiKey.UserId == 0 {
			continue
		}
		// Add '*' only in the case where we are printing out tables
		outputKey := apiKey.Key
		if outputWriter.GetOutputFormat() == output.Human {
			if clusterId != "" && apiKey.Key == currentKey {
				outputKey = fmt.Sprintf("* %s", apiKey.Key)
			} else {
				outputKey = fmt.Sprintf("  %s", apiKey.Key)
			}
		}

		email := c.getEmail(apiKey, serviceAccountsMap, usersMap)

		created := time.Unix(apiKey.Created.Seconds, 0).In(time.UTC).Format(time.RFC3339)
		userResourceId := apiKey.UserResourceId
		if userResourceId == "" {
			userResourceId = resourceIdMap[apiKey.UserId]
		}
		// If resource id is empty then the resource was not specified, or Cloud was specified.
		// Note that if more resource types are added with no logical clusters, then additional logic
		// needs to be added here to determine the resource type.
		if clusterId == "" && len(apiKey.LogicalClusters) == 0 {
			// Cloud key.
			outputWriter.AddElement(&keyDisplay{
				Key:            outputKey,
				Description:    apiKey.Description,
				UserResourceId: userResourceId,
				UserEmail:      email,
				ResourceType:   resource.Cloud,
				Created:        created,
			})
		}

		if resourceType == resource.Cloud {
			continue
		}

		for _, lc := range apiKey.LogicalClusters {
			outputWriter.AddElement(&keyDisplay{
				Key:            outputKey,
				Description:    apiKey.Description,
				UserResourceId: userResourceId,
				UserEmail:      email,
				ResourceType:   lc.Type,
				ResourceId:     lc.Id,
				Created:        created,
			})
		}
	}

	return outputWriter.Out()
}

func getServiceAccountsMap(serviceAccounts []*orgv1.User) map[int32]bool {
	saMap := make(map[int32]bool)
	for _, sa := range serviceAccounts {
		saMap[sa.Id] = true
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

func (c *command) mapUserIdToResourceId(users []*orgv1.User) map[int32]string {
	idMap := make(map[int32]string)
	for _, user := range users {
		idMap[user.Id] = user.ResourceId
	}
	return idMap
}

func (c *command) mapResourceIdToUserId(users []*orgv1.User) map[string]int32 {
	idMap := make(map[string]int32)
	for _, user := range users {
		idMap[user.ResourceId] = user.Id
	}
	return idMap
}

func (c *command) getEmail(apiKey *schedv1.ApiKey, serviceAccountsMap map[int32]bool, usersMap map[int32]*orgv1.User) string {
	if _, ok := serviceAccountsMap[apiKey.UserId]; ok {
		return "<service account>"
	}

	if auditLog, ok := pcmd.AreAuditLogsEnabled(c.State); ok && auditLog.ServiceAccountId == apiKey.UserId {
		return "<auditlog service account>"
	}

	if user, ok := usersMap[apiKey.UserId]; ok {
		return user.Email
	}

	// check if api key is owned by current user
	if c.State.Auth.User.Id == apiKey.UserId {
		return c.State.Auth.User.Email
	}

	return "<deactivated user>"
}
