package apikey

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	configv1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/keystore"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/shell/completer"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const longDescription = `Use this command to register an API secret created by another
process and store it locally.

When you create an API key with the CLI, it is automatically stored locally.
However, when you create an API key using the UI, API, or with the CLI on another
machine, the secret is not available for CLI use until you "store" it. This is because
secrets are irretrievable after creation.

You must have an API secret stored locally for certain CLI commands to
work. For example, the Kafka topic consume and produce commands require an API secret.

There are five ways to pass the secret:
1. api-key store <key> <secret>.
2. api-key store; you will be prompted for both API key and secret.
3. api-key store <key>; you will be prompted for API secret.
4. api-key store <key> -; for piping API secret.
5. api-key store <key> @<filepath>.
`

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
	keystore                keystore.KeyStore
	flagResolver            pcmd.FlagResolver
	completableChildren     []*cobra.Command
	completableFlagChildren map[string][]*cobra.Command
	analyticsClient         analytics.Client
}

var (
	listFields              = []string{"Key", "Description", "UserId", "UserResourceId", "UserEmail", "ResourceType", "ResourceId", "Created"}
	listHumanLabels         = []string{"Key", "Description", "Owner", "Owner Resource Id", "Owner Email", "Resource Type", "Resource ID", "Created"}
	listStructuredLabels    = []string{"key", "description", "owner", "owner_resource_id", "owner_email", "resource_type", "resource_id", "created"}
	createFields            = []string{"Key", "Secret"}
	createHumanRenames      = map[string]string{"Key": "API Key"}
	createStructuredRenames = map[string]string{"Key": "key", "Secret": "secret"}
	resourceFlagName        = "resource"
)

// New returns the Cobra command for API Key.
func New(prerunner pcmd.PreRunner, keystore keystore.KeyStore, resolver pcmd.FlagResolver, analyticsClient analytics.Client) *command {
	cliCmd := pcmd.NewAuthenticatedStateFlagCommand(
		&cobra.Command{
			Use:         "api-key",
			Short:       "Manage the API keys.",
			Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		}, prerunner, SubcommandFlags)
	cmd := &command{
		AuthenticatedStateFlagCommand: cliCmd,
		keystore:                      keystore,
		flagResolver:                  resolver,
		analyticsClient:               analyticsClient,
	}
	cmd.init()
	return cmd
}

func (c *command) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List the API keys.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List the API keys that belong to service account with resource ID `sa-lqv3mm` on cluster `lkc-xyz`",
				Code: `confluent api-key list --resource lkc-xyz --service-account sa-lqv3mm `,
			},
		),
	}
	listCmd.Flags().String(resourceFlagName, "", "The resource ID to filter by. Use \"cloud\" to show only Cloud API keys.")
	listCmd.Flags().Bool("current-user", false, "Show only API keys belonging to current user.")
	listCmd.Flags().String("service-account", "", "The service account ID to filter by.")
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create API keys for a given resource.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create an API key for service account with resource ID `sa-lqv3mm` for cluster `lkc-xyz`",
				Code: `confluent api-key create --resource lkc-xyz --service-account sa-lqv3mm`,
			},
		),
	}
	createCmd.Flags().String(resourceFlagName, "", "The resource ID. Use \"cloud\" to create a Cloud API key.")
	createCmd.Flags().String("service-account", "", "Service account ID. If not specified, the API key will have full access on the cluster.")
	createCmd.Flags().String("description", "", "Description of API key.")
	createCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	createCmd.Flags().SortFlags = false
	if err := createCmd.MarkFlagRequired(resourceFlagName); err != nil {
		panic(err)
	}
	c.AddCommand(createCmd)

	updateCmd := &cobra.Command{
		Use:   "update <apikey>",
		Short: "Update an API key.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.update),
	}
	updateCmd.Flags().String("description", "", "Description of the API key.")
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <apikey>",
		Short: "Delete an API key.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.delete),
	}
	c.AddCommand(deleteCmd)

	storeCmd := &cobra.Command{
		Use:   "store <apikey> <secret>",
		Short: "Store an API key/secret locally to use in the CLI.",
		Long:  longDescription,
		Args:  cobra.MaximumNArgs(2),
		RunE:  pcmd.NewCLIRunE(c.store),
	}
	storeCmd.Flags().String(resourceFlagName, "", "The resource ID of the resource the API key is for.")
	storeCmd.Flags().BoolP("force", "f", false, "Force overwrite existing secret for this key.")
	storeCmd.Flags().SortFlags = false
	c.AddCommand(storeCmd)

	useCmd := &cobra.Command{
		Use:   "use <apikey>",
		Short: "Make an API key active for use in other commands.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.use),
	}
	useCmd.Flags().String(resourceFlagName, "", "The resource ID.")
	useCmd.Flags().SortFlags = false
	if err := useCmd.MarkFlagRequired(resourceFlagName); err != nil {
		panic(err)
	}
	c.AddCommand(useCmd)
	c.completableChildren = append(c.completableChildren, updateCmd, deleteCmd, storeCmd, useCmd)
	c.completableFlagChildren = map[string][]*cobra.Command{
		resourceFlagName:  {createCmd, listCmd, storeCmd, useCmd},
		"service-account": {createCmd},
	}
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	c.setKeyStoreIfNil()
	type keyDisplay struct {
		Key            string
		Description    string
		UserId         int32
		UserResourceId string
		UserEmail      string
		ResourceType   string
		ResourceId     string
		Created        string
	}
	var apiKeys []*schedv1.ApiKey

	resourceType, resourceId, currentKey, err := c.resolveResourceId(cmd, c.Config.Resolver, c.Client)
	if err != nil {
		return err
	}
	var logicalClusters []*schedv1.ApiKey_Cluster
	if resourceId != "" {
		logicalClusters = []*schedv1.ApiKey_Cluster{{Id: resourceId, Type: resourceType}}
	}

	serviceAccountId, err := cmd.Flags().GetString("service-account")
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
	allUsers := append(serviceAccounts, users...)

	userId := int32(0)
	serviceAccount := false
	if serviceAccountId != "" {
		serviceAccount = true
		if isResourceId(serviceAccountId) { // if user inputs resource ID, get corresponding numeric ID
			userIdMap, err := c.mapResourceIdToUserId(allUsers)
			if err != nil {
				return err
			}
			var ok bool
			userId, ok = userIdMap[serviceAccountId]
			if !ok {
				return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.ServiceAccountNotFoundErrorMsg, serviceAccountId), errors.ServiceAccountNotFoundSuggestions)
			}
		} else { // if user inputs numeric ID, convert it to int32
			userIdp, _ := strconv.Atoi(serviceAccountId)
			userId = int32(userIdp)
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
	resourceIdMap, err := c.mapUserIdToResourceId(allUsers)
	if err != nil {
		return err
	}

	for _, apiKey := range apiKeys {
		// ignore keys owned by Confluent-internal user (healthcheck, etc)
		if apiKey.UserId == 0 {
			continue
		}
		// Add '*' only in the case where we are printing out tables
		outputKey := apiKey.Key
		if outputWriter.GetOutputFormat() == output.Human {
			if resourceId != "" && apiKey.Key == currentKey {
				outputKey = fmt.Sprintf("* %s", apiKey.Key)
			} else {
				outputKey = fmt.Sprintf("  %s", apiKey.Key)
			}
		}

		var email string
		if _, ok := serviceAccountsMap[apiKey.UserId]; ok {
			email = "<service account>"
		} else {
			auditLog, enabled := pcmd.IsAuditLogsEnabled(c.State)
			if enabled && auditLog.ServiceAccountId == apiKey.UserId {
				email = "<auditlog service account>"
			} else {
				if user, ok := usersMap[apiKey.UserId]; ok {
					email = user.Email
				} else {
					email = "<deactivated user>"
				}
			}
		}

		created := time.Unix(apiKey.Created.Seconds, 0).In(time.UTC).Format(time.RFC3339)
		userResourceId := apiKey.UserResourceId
		if userResourceId == "" {
			userResourceId = resourceIdMap[apiKey.UserId]
		}
		// If resource id is empty then the resource was not specified, or Cloud was specified.
		// Note that if more resource types are added with no logical clusters, then additional logic
		// needs to be added here to determine the resource type.
		if resourceId == "" && len(apiKey.LogicalClusters) == 0 {
			// Cloud key.
			outputWriter.AddElement(&keyDisplay{
				Key:            outputKey,
				Description:    apiKey.Description,
				UserId:         apiKey.UserId,
				UserResourceId: userResourceId,
				UserEmail:      email,
				ResourceType:   pcmd.CloudResourceType,
				Created:        created,
			})
		}

		if resourceType == pcmd.CloudResourceType {
			continue
		}

		for _, lc := range apiKey.LogicalClusters {
			outputWriter.AddElement(&keyDisplay{
				Key:            outputKey,
				Description:    apiKey.Description,
				UserId:         apiKey.UserId,
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

func (c *command) update(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()
	apiKey := args[0]
	key, err := c.Client.APIKey.Get(context.Background(), &schedv1.ApiKey{Key: apiKey, AccountId: c.EnvironmentId()})
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("description") {
		key.Description = description
	}

	err = c.Client.APIKey.Update(context.Background(), key)
	if err != nil {
		return err
	}
	if cmd.Flags().Changed("description") {
		utils.ErrPrintf(cmd, errors.UpdateSuccessMsg, "description", "API key", apiKey, description)
	}
	return nil
}

func (c *command) create(cmd *cobra.Command, _ []string) error {
	c.setKeyStoreIfNil()
	resourceType, clusterId, _, err := c.resolveResourceId(cmd, c.Config.Resolver, c.Client)
	if err != nil {
		return err
	}
	serviceAccountId, err := cmd.Flags().GetString("service-account")
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	key := &schedv1.ApiKey{
		Description: description,
		AccountId:   c.EnvironmentId(),
	}

	key, err = c.completeKeyId(key, serviceAccountId) // get corresponding numeric/resource ID if the cmd has a service-account flag
	if err != nil {
		return err
	}

	if resourceType != pcmd.CloudResourceType {
		key.LogicalClusters = []*schedv1.ApiKey_Cluster{{Id: clusterId, Type: resourceType}}
	}
	userKey, err := c.Client.APIKey.Create(context.Background(), key)
	if err != nil {
		return err
	}

	outputFormat, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	if outputFormat == output.Human.String() {
		utils.ErrPrintln(cmd, errors.APIKeyTime)
		utils.ErrPrintln(cmd, errors.APIKeyNotRetrievableMsg)
	}

	err = output.DescribeObject(cmd, userKey, createFields, createHumanRenames, createStructuredRenames)
	if err != nil {
		return err
	}

	if resourceType == pcmd.KafkaResourceType {
		if err := c.keystore.StoreAPIKey(userKey, clusterId, cmd); err != nil {
			return errors.Wrap(err, errors.UnableToStoreAPIKeyErrorMsg)
		}
	}

	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, userKey.UserResourceId)
	c.analyticsClient.SetSpecialProperty(analytics.ApiKeyPropertiesKey, userKey.Key)
	return nil
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()
	apiKey := args[0]

	userKey, err := c.Client.APIKey.Get(context.Background(), &schedv1.ApiKey{Key: apiKey, AccountId: c.EnvironmentId()})
	if err != nil {
		return err
	}
	key := &schedv1.ApiKey{
		Id:             userKey.Id,
		Key:            apiKey,
		AccountId:      c.EnvironmentId(),
		UserId:         userKey.UserId,
		UserResourceId: userKey.UserResourceId,
	}

	err = c.Client.APIKey.Delete(context.Background(), key)
	if err != nil {
		return err
	}
	utils.Printf(cmd, errors.DeletedAPIKeyMsg, apiKey)
	err = c.keystore.DeleteAPIKey(apiKey)
	if err != nil {
		return err
	}
	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, key.UserResourceId)
	c.analyticsClient.SetSpecialProperty(analytics.ApiKeyPropertiesKey, key.Key)
	return nil
}

func (c *command) store(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()

	var cluster *configv1.KafkaClusterConfig

	// Attempt to get cluster from --resource flag if set; if that doesn't work,
	// attempt to fall back to the currently active Kafka cluster
	resourceType, clusterId, _, err := c.resolveResourceId(cmd, c.Config.Resolver, c.Client)
	if err == nil && clusterId != "" {
		if resourceType != pcmd.KafkaResourceType {
			return errors.Errorf(errors.NonKafkaNotImplementedErrorMsg)
		}
		cluster, err = c.Context.FindKafkaCluster(cmd, clusterId)
		if err != nil {
			return err
		}
	} else {
		cluster, err = c.Context.GetKafkaClusterForCommand(cmd)
		if err != nil {
			return err
		}
	}

	var key string
	if len(args) == 0 {
		key, err = c.parseFlagResolverPromptValue("", "Key: ", false)
		if err != nil {
			return err
		}
	} else {
		key = args[0]
	}

	var secret string
	if len(args) < 2 {
		secret, err = c.parseFlagResolverPromptValue("", "Secret: ", true)
		if err != nil {
			return err
		}
	} else if len(args) == 2 {
		secret, err = c.parseFlagResolverPromptValue(args[1], "", true)
		if err != nil {
			return err
		}
	}
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	// Check if API key exists server-side
	apiKey, err := c.Client.APIKey.Get(context.Background(), &schedv1.ApiKey{Key: key, AccountId: c.EnvironmentId()})
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), errors.APIKeyNotFoundSuggestions)
	}

	apiKeyIsValidForTargetCluster := false
	for _, lkc := range apiKey.LogicalClusters {
		if lkc.Id == cluster.ID {
			apiKeyIsValidForTargetCluster = true
			break
		}
	}
	if !apiKeyIsValidForTargetCluster {
		return errors.NewErrorWithSuggestions(errors.APIKeyNotValidForClusterErrorMsg, errors.APIKeyNotValidForClusterSuggestions)
	}

	// API key exists server-side... now check if API key exists locally already
	if found, err := c.keystore.HasAPIKey(key, cluster.ID, cmd); err != nil {
		return err
	} else if found && !force {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.RefuseToOverrideSecretErrorMsg, key),
			fmt.Sprintf(errors.RefuseToOverrideSecretSuggestions, key))
	}

	if err := c.keystore.StoreAPIKey(&schedv1.ApiKey{Key: key, Secret: secret}, cluster.ID, cmd); err != nil {
		return errors.Wrap(err, errors.UnableToStoreAPIKeyErrorMsg)
	}
	utils.ErrPrintf(cmd, errors.StoredAPIKeyMsg, key)
	return nil
}

func (c *command) use(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()
	apiKey := args[0]
	resourceType, clusterId, _, err := c.resolveResourceId(cmd, c.Config.Resolver, c.Client)
	if err != nil {
		return err
	}
	if resourceType != pcmd.KafkaResourceType {
		return errors.Errorf(errors.NonKafkaNotImplementedErrorMsg)
	}
	cluster, err := c.Context.FindKafkaCluster(cmd, clusterId)
	if err != nil {
		return err
	}
	err = c.Context.UseAPIKey(cmd, apiKey, cluster.ID)
	if err != nil {
		return errors.NewWrapErrorWithSuggestions(err, errors.APIKeyUseFailedErrorMsg, fmt.Sprintf(errors.APIKeyUseFailedSuggestions, apiKey))
	}
	utils.Printf(cmd, errors.UseAPIKeyMsg, apiKey, clusterId)
	return nil
}

func (c *command) setKeyStoreIfNil() {
	if c.keystore == nil {
		c.keystore = &keystore.ConfigKeyStore{Config: c.Config}
	}
}

func (c *command) parseFlagResolverPromptValue(source, prompt string, secure bool) (string, error) {
	val, err := c.flagResolver.ValueFrom(source, prompt, secure)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(val), nil
}

// Completable implementation

func (c *command) Cmd() *cobra.Command {
	return c.Command
}

func (c *command) ServerComplete() []prompt.Suggest {
	var suggests []prompt.Suggest
	apiKeys, err := c.fetchAPIKeys()
	if err != nil {
		return suggests
	}
	for _, key := range apiKeys {
		suggests = append(suggests, prompt.Suggest{
			Text:        key.Key,
			Description: key.Description,
		})
	}
	return suggests
}

func (c *command) fetchAPIKeys() ([]*schedv1.ApiKey, error) {
	apiKeys, err := c.Client.APIKey.List(context.Background(), &schedv1.ApiKey{AccountId: c.EnvironmentId(), LogicalClusters: nil, UserId: 0})
	if err != nil {
		return nil, errors.HandleCommon(err, c.Command)
	}

	var userApiKeys []*schedv1.ApiKey
	for _, key := range apiKeys {
		if key.UserId != 0 {
			userApiKeys = append(userApiKeys, key)
		}
	}
	return userApiKeys, nil
}

func (c *command) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}

func (c *command) ServerCompletableFlagChildren() map[string][]*cobra.Command {
	return c.completableFlagChildren
}

func (c *command) ServerFlagComplete() map[string]func() []prompt.Suggest {
	return map[string]func() []prompt.Suggest{
		resourceFlagName:  c.resourceFlagCompleterFunc,
		"service-account": completer.ServiceAccountFlagCompleterFunc(c.Client),
	}
}

func (c *command) resourceFlagCompleterFunc() []prompt.Suggest {
	suggestions := completer.ClusterFlagServerCompleterFunc(c.Client, c.EnvironmentId())()

	ctx := context.Background()
	ctxClient := pcmd.NewContextClient(c.Context)
	cluster, err := ctxClient.FetchSchemaRegistryByAccountId(ctx, c.EnvironmentId())
	if err == nil {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        cluster.Id,
			Description: cluster.Name,
		})
	}
	req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId()}
	clusters, err := c.Client.KSQL.List(context.Background(), req)
	if err == nil {
		for _, cluster := range clusters {
			suggestions = append(suggestions, prompt.Suggest{
				Text:        cluster.Id,
				Description: cluster.Name,
			})
		}
	}
	return suggestions
}

func (c *command) getAllUsers() ([]*orgv1.User, error) {
	serviceAccounts, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return nil, err
	}
	adminUsers, err := c.Client.User.List(context.Background())
	if err != nil {
		return nil, err
	}
	return append(serviceAccounts, adminUsers...), nil
}

func (c *command) completeKeyId(key *schedv1.ApiKey, Id string) (*schedv1.ApiKey, error) {
	if Id != "" { // it has a service-account flag
		key.ServiceAccount = true
		users, err := c.getAllUsers()
		if err != nil {
			return key, err
		}
		idp, err := strconv.Atoi(Id)
		if err != nil { // it's a resource id
			key.UserResourceId = Id
			for _, user := range users {
				if Id == user.ResourceId {
					key.UserId = user.Id
				}
			}
		} else { // it's a numeric id
			key.UserId = int32(idp)
			for _, user := range users {
				if int32(idp) == user.Id {
					key.UserResourceId = user.ResourceId
				}
			}
		}
	} else {
		key.ServiceAccount = false
	}
	return key, nil
}

func (c *command) mapUserIdToResourceId(users []*orgv1.User) (map[int32]string, error) {
	idMap := make(map[int32]string)
	for _, user := range users {
		idMap[user.Id] = user.ResourceId
	}
	return idMap, nil
}

func (c *command) mapResourceIdToUserId(users []*orgv1.User) (map[string]int32, error) {
	idMap := make(map[string]int32)
	for _, user := range users {
		idMap[user.ResourceId] = user.Id
	}
	return idMap, nil
}

func isResourceId(Id string) bool {
	_, err := strconv.Atoi(Id)
	return err != nil
}
