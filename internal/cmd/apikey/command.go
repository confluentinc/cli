package apikey

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/keystore"
	"github.com/confluentinc/cli/internal/pkg/output"
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
	*pcmd.AuthenticatedCLICommand
	keystore     keystore.KeyStore
	flagResolver pcmd.FlagResolver
}

var (
	listFields              = []string{"Key", "UserId", "Description", "ResourceType", "ResourceId"}
	listHumanLabels         = []string{"Key", "Owner", "Description", "Resource Type", "Resource ID"}
	listStructuredLabels    = []string{"key", "owner", "description", "resource_type", "resource_id"}
	createFields            = []string{"Key", "Secret"}
	createHumanRenames      = map[string]string{"Key": "API Key"}
	createStructuredRenames = map[string]string{"Key": "key", "Secret": "secret"}
	resourceFlagName        = "resource"
)

// New returns the Cobra command for API Key.
func New(prerunner pcmd.PreRunner, keystore keystore.KeyStore, resolver pcmd.FlagResolver) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "api-key",
			Short: "Manage the API keys.",
		}, prerunner)
	cmd := &command{
		AuthenticatedCLICommand: cliCmd,
		keystore:                keystore,
		flagResolver:            resolver,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List the API keys.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}
	listCmd.Flags().String(resourceFlagName, "", "The resource ID to filter by. Use \"cloud\" to show only Cloud API keys.")
	listCmd.Flags().Bool("current-user", false, "Show only API keys belonging to current user.")
	listCmd.Flags().Int32("service-account", 0, "The service account ID to filter by.")
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create API keys for a given resource.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.create),
	}
	createCmd.Flags().String(resourceFlagName, "", "The resource ID. Use \"cloud\" to create a Cloud API key.")
	createCmd.Flags().Int32("service-account", 0, "Service account ID. If not specified, the API key will have full access on the cluster.")
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

	c.AddCommand(&cobra.Command{
		Use:   "delete <apikey>",
		Short: "Delete an API key.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.delete),
	})

	storeCmd := &cobra.Command{
		Use:   "store <apikey> <secret>",
		Short: "Store an API key/secret locally to use in the CLI.",
		Long:  longDescription,
		Args:  cobra.MaximumNArgs(2),
		RunE:  pcmd.NewCLIRunE(c.store),
	}
	storeCmd.Flags().String(resourceFlagName, "", "The resource ID.")
	storeCmd.Flags().BoolP("force", "f", false, "Force overwrite existing secret for this key.")
	storeCmd.Flags().SortFlags = false
	if err := storeCmd.MarkFlagRequired(resourceFlagName); err != nil {
		panic(err)
	}
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
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	c.setKeyStoreIfNil()
	type keyDisplay struct {
		Key          string
		Description  string
		UserId       int32
		ResourceType string
		ResourceId   string
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

	userId, err := cmd.Flags().GetInt32("service-account")
	if err != nil {
		return err
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

	apiKeys, err = c.Client.APIKey.List(context.Background(), &schedv1.ApiKey{AccountId: c.EnvironmentId(), LogicalClusters: logicalClusters, UserId: userId})
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}

	for _, apiKey := range apiKeys {
		// ignore keys owned by Confluent-internal user (healthcheck, etc)
		if apiKey.UserId == 0 {
			continue
		}
		// Add '*' only in the case where we are printing out tables
		if outputWriter.GetOutputFormat() == output.Human {
			if resourceId != "" && apiKey.Key == currentKey {
				apiKey.Key = fmt.Sprintf("* %s", apiKey.Key)
			} else {
				apiKey.Key = fmt.Sprintf("  %s", apiKey.Key)
			}
		}

		// If resource id is empty then the resource was not specified, or Cloud was specified.
		// Note that if more resource types are added with no logical clusters, then additional logic
		// needs to be added here to determine the resource type.
		if resourceId == "" && len(apiKey.LogicalClusters) == 0 {
			// Cloud key.
			outputWriter.AddElement(&keyDisplay{
				Key:          apiKey.Key,
				Description:  apiKey.Description,
				UserId:       apiKey.UserId,
				ResourceType: pcmd.CloudResourceType,
			})
		}

		if resourceType == pcmd.CloudResourceType {
			continue
		}

		for _, lc := range apiKey.LogicalClusters {
			outputWriter.AddElement(&keyDisplay{
				Key:          apiKey.Key,
				Description:  apiKey.Description,
				UserId:       apiKey.UserId,
				ResourceType: lc.Type,
				ResourceId:   lc.Id,
			})
		}
	}
	return outputWriter.Out()
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
		pcmd.ErrPrintf(cmd, errors.UpdateSuccessMsg, "description", "API key", apiKey, description)
	}
	return nil
}

func (c *command) create(cmd *cobra.Command, _ []string) error {
	c.setKeyStoreIfNil()
	resourceType, clusterId, _, err := c.resolveResourceId(cmd, c.Config.Resolver, c.Client)
	if err != nil {
		return err
	}
	userId, err := cmd.Flags().GetInt32("service-account")
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	key := &schedv1.ApiKey{
		UserId:      userId,
		Description: description,
		AccountId:   c.EnvironmentId(),
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
		pcmd.ErrPrintln(cmd, errors.APIKeyTime)
		pcmd.ErrPrintln(cmd, errors.APIKeyNotRetrievableMsg)
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
		Id:        userKey.Id,
		Key:       apiKey,
		AccountId: c.EnvironmentId(),
	}

	err = c.Client.APIKey.Delete(context.Background(), key)
	if err != nil {
		return err
	}
	pcmd.Printf(cmd, errors.DeletedAPIKeyMsg, apiKey)
	return c.keystore.DeleteAPIKey(apiKey, cmd)
}

func (c *command) store(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()

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
	_, err = c.Client.APIKey.Get(context.Background(), &schedv1.ApiKey{Key: key, AccountId: c.EnvironmentId()})
	if err != nil {
		return err
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
	pcmd.ErrPrintf(cmd, errors.StoredAPIKeyMsg, key)
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
		return errors.Wrap(err, errors.APIKeyUseFailedErrorMsg)
	}
	pcmd.Printf(cmd, errors.UseAPIKeyMsg, apiKey, clusterId)
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
