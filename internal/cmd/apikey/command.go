package apikey

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	"github.com/confluentinc/go-printer"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/keystore"
)

const longDescription = `Use this command to register an API secret created by another
process and store it locally.

When you create an API key with the CLI, it is automatically stored locally.
However, when you create an API key using the UI, API, or with the CLI on another
machine, the secret is not available for CLI use until you "store" it. This is because
secrets are irretrievable after creation.

You must have an API secret stored locally for certain CLI commands to
work. For example, the Kafka topic consume and produce commands require an API secret.
`

type command struct {
	*cobra.Command
	config   *config.Config
	client   ccloud.APIKey
	ch       *pcmd.ConfigHelper
	keystore keystore.KeyStore
}

var (
	listFields    = []string{"Key", "UserId", "Description", "ResourceType", "ResourceId"}
	listLabels    = []string{"Key", "Owner", "Description", "Resource Type", "Resource ID"}
	createFields  = []string{"Key", "Secret"}
	createRenames = map[string]string{"Key": "API Key"}
)

// New returns the Cobra command for API Key.
func New(prerunner pcmd.PreRunner, config *config.Config, client ccloud.APIKey, ch *pcmd.ConfigHelper, keystore keystore.KeyStore) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:               "api-key",
			Short:             "Manage the API keys.",
			PersistentPreRunE: prerunner.Authenticated(),
		},
		config:   config,
		client:   client,
		ch:       ch,
		keystore: keystore,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List the API keys.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	}
	listCmd.Flags().String("resource", "", "The resource ID.")
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create API keys for a given resource.",
		RunE:  c.create,
		Args:  cobra.NoArgs,
	}
	createCmd.Flags().String("resource", "", "The resource ID.")
	createCmd.Flags().Int32("service-account-id", 0, "Service account ID. If not specified, the API key will have full access on the cluster.")
	createCmd.Flags().String("description", "", "Description of API key.")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	updateCmd := &cobra.Command{
		Use:   "update <apikey>",
		Short: "Update API key.",
		RunE:  c.update,
		Args:  cobra.ExactArgs(1),
	}
	updateCmd.Flags().String("description", "", "Description of the API key.")
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)

	c.AddCommand(&cobra.Command{
		Use:   "delete <apikey>",
		Short: "Delete API keys.",
		RunE:  c.delete,
		Args:  cobra.ExactArgs(1),
	})

	storeCmd := &cobra.Command{
		Use:   "store <apikey> <secret>",
		Short: `Store an API key/secret locally to use in the CLI.`,
		Long:  longDescription,
		RunE:  c.store,
		Args:  cobra.ExactArgs(2),
	}
	storeCmd.Flags().String("resource", "", "The resource ID.")
	storeCmd.Flags().BoolP("force", "f", false, "Force overwrite existing secret for this key.")
	storeCmd.Flags().SortFlags = false
	c.AddCommand(storeCmd)

	useCmd := &cobra.Command{
		Use:   "use <apikey>",
		Short: "Make API key active for use in other commands.",
		RunE:  c.use,
		Args:  cobra.ExactArgs(1),
	}
	useCmd.Flags().String("resource", "", "The resource ID.")
	useCmd.Flags().SortFlags = false
	c.AddCommand(useCmd)
}

func (c *command) list(cmd *cobra.Command, args []string) error {
	type keyDisplay struct {
		Key          string
		Description  string
		UserId       int32
		ResourceType string
		ResourceId   string
	}
	var apiKeys []*authv1.ApiKey
	var data [][]string

	resourceType, accId, resourceId, currentKey, err := c.resolveResourceID(cmd, args)
	//Return resource not found errors
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	apiKeys, err = c.client.List(context.Background(), &authv1.ApiKey{AccountId: accId, LogicalClusters: []*authv1.ApiKey_Cluster{{Id: resourceId, Type: resourceType}}})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	for _, apiKey := range apiKeys {
		// ignore keys owned by Confluent-internal user (healthcheck, etc)
		if apiKey.UserId == 0 {
			continue
		}

		if apiKey.Key == currentKey {
			apiKey.Key = fmt.Sprintf("* %s", apiKey.Key)
		} else {
			apiKey.Key = fmt.Sprintf("  %s", apiKey.Key)
		}

		for _, c := range apiKey.LogicalClusters {
			if c.Id == resourceId {
				data = append(data, printer.ToRow(&keyDisplay{
					Key:          apiKey.Key,
					Description:  apiKey.Description,
					UserId:       apiKey.UserId,
					ResourceType: resourceType,
					ResourceId:   resourceId,
				}, listFields))
				break
			}
		}
	}
	printer.RenderCollectionTable(data, listLabels)
	return nil
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	apiKey := args[0]

	key, err := c.client.Get(context.Background(), &authv1.ApiKey{Key: apiKey, AccountId: c.config.Auth.Account.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if cmd.Flags().Changed("description") {
		key.Description = description
	}

	err = c.client.Update(context.Background(), key)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return nil
}

func (c *command) create(cmd *cobra.Command, args []string) error {

	resourceType, accId, clusterId, _, err := c.resolveResourceID(cmd, args)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	environment, err := pcmd.GetEnvironment(cmd, c.config)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	userId, err := cmd.Flags().GetInt32("service-account-id")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	key := &authv1.ApiKey{
		UserId:          userId,
		Description:     description,
		AccountId:       accId,
		LogicalClusters: []*authv1.ApiKey_Cluster{{Id: clusterId, Type: resourceType}},
	}

	userKey, err := c.client.Create(context.Background(), key)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	pcmd.Println(cmd, "Save the API key and secret. The secret is not retrievable later.")
	err = printer.RenderTableOut(userKey, createFields, createRenames, os.Stdout)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if resourceType == kafkaResourceType {
		if err := c.keystore.StoreAPIKey(userKey, clusterId, environment); err != nil {
			return errors.HandleCommon(errors.Wrapf(err, "Unable to store API key locally."), cmd)
		}
	}
	return nil
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	apiKey := args[0]

	userKey, err := c.client.Get(context.Background(), &authv1.ApiKey{Key: apiKey, AccountId: c.config.Auth.Account.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	key := &authv1.ApiKey{
		Id:        userKey.Id,
		Key:       apiKey,
		AccountId: c.config.Auth.Account.Id,
	}

	err = c.client.Delete(context.Background(), key)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Println(cmd, "API Key successfully deleted.")
	return c.keystore.DeleteAPIKey(apiKey)
}

func (c *command) store(cmd *cobra.Command, args []string) error {
	key := args[0]
	secret := args[1]

	kcc, err := pcmd.GetKafkaClusterConfig(cmd, c.ch, "resource")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	environment, err := pcmd.GetEnvironment(cmd, c.config)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	// Check if API key exists server-side
	_, err = c.client.Get(context.Background(), &authv1.ApiKey{Key: key, AccountId: c.config.Auth.Account.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	// API key exists server-side... now check if API key exists locally already
	if found, err := c.keystore.HasAPIKey(key, kcc.ID, environment); err != nil {
		return errors.HandleCommon(err, cmd)
	} else if found && !force {
		return errors.HandleCommon(errors.Errorf("Refusing to overwrite existing secret for API Key %s", key), cmd)
	}

	if err := c.keystore.StoreAPIKey(&authv1.ApiKey{Key: key, Secret: secret}, kcc.ID, environment); err != nil {
		return errors.HandleCommon(errors.Wrapf(err, "Unable to store the API key locally."), cmd)
	}
	return nil
}

func (c *command) use(cmd *cobra.Command, args []string) error {
	apiKey := args[0]

	cluster, err := pcmd.GetKafkaCluster(cmd, c.ch, "resource")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	err = c.ch.UseAPIKey(apiKey, cluster.Id)
	if err != nil {
		// This will error if no secret is stored
		return errors.HandleCommon(err, cmd)
	}
	return nil
}
