package apikey

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	"github.com/confluentinc/go-printer"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

type command struct {
	*cobra.Command
	config *config.Config
	client ccloud.APIKey
	kafka  ccloud.Kafka
}

var (
	listFields    = []string{"Key", "UserId"}
	listLabels    = []string{"Key", "Owner"}
	createFields  = []string{"Key", "Secret"}
	createRenames = map[string]string{"Key": "API Key"}
)

// New returns the Cobra command for API Key.
func New(prerunner pcmd.PreRunner, config *config.Config, client ccloud.APIKey, kafka ccloud.Kafka) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:               "api-key",
			Short:             "Manage API keys",
			PersistentPreRunE: prerunner.Authenticated(),
		},
		config: config,
		client: client,
		kafka:  kafka,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.PersistentFlags().String("environment", "", "ID of the environment in which to run the command")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List API keys",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	}
	listCmd.Flags().String("cluster", "", "Cluster ID to list API keys for")
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)


	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create API key",
		RunE:  c.create,
		Args:  cobra.NoArgs,
	}
	createCmd.Flags().String("cluster", "", "Grant access to a cluster with this ID")
	createCmd.Flags().Int32("service-account-id", 0, "Create API key for a service account")
	createCmd.Flags().String("description", "", "Description or purpose for the API key")
	createCmd.Flags().Bool("use", false, "Activate this API key for the cluster")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	c.AddCommand(&cobra.Command{
		Use:   "delete KEY",
		Short: "Delete API key",
		RunE:  c.delete,
		Args:  cobra.ExactArgs(1),
	})

	useCmd := &cobra.Command{
		Use:   "use",
		Short: "Make the API key active for use in other commands",
		RunE:  c.use,
		Args:  cobra.ExactArgs(1),
	}
	useCmd.Flags().String("cluster", "", "Make this API key active for this cluster")
	// TODO: cluster (flag and active) handling will be cleaned up as part of CLI-112
	// In PR: https://github.com/confluentinc/cli/pull/146/files#diff-7bf5d7c832065ed38ccd25c6c525b13bR148
	_ = useCmd.MarkFlagRequired("cluster")
	c.AddCommand(useCmd)
}

func (c *command) list(cmd *cobra.Command, args []string) error {
	apiKeys, err := c.client.List(context.Background(), &authv1.ApiKey{AccountId: c.config.Auth.Account.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	type keyDisplay struct {
		Key         string
		Description string
		UserId      int32
	}

	ctx, err := c.config.Context()
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	var data [][]string
	for _, apiKey := range apiKeys {
		// ignore keys owned by Confluent-internal user (healthcheck, etc)
		if apiKey.UserId == 0 {
			continue
		}

		for _, c := range apiKey.LogicalClusters {
			if c.Id == ctx.Kafka {
				data = append(data, printer.ToRow(&keyDisplay{
					Key:         apiKey.Key,
					Description: apiKey.Description,
					UserId:      apiKey.UserId,
				}, listFields))
				break
			}
		}
	}

	printer.RenderCollectionTable(data, listLabels)
	return nil
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.GetKafkaCluster(cmd, c.config)
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

	use, err := cmd.Flags().GetBool("use")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	environment, err := pcmd.GetEnvironment(cmd, c.config)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	key := &authv1.ApiKey{
		UserId:          userId,
		Description:     description,
		AccountId:       c.config.Auth.Account.Id,
		LogicalClusters: []*authv1.ApiKey_Cluster{{Id: cluster.Id}},
	}

	userKey, err := c.client.Create(context.Background(), key)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	pcmd.Println(cmd, "Please save the API Key and Secret. THIS IS THE ONLY CHANCE YOU HAVE!")
	err = printer.RenderTableOut(userKey, createFields, createRenames, os.Stdout)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if err := c.createKey(userKey, environment, clusterID); err != nil {
		return errors.HandleCommon(errors.Wrapf(err, "unable to save api key locally"), cmd)
	}

	if use {
		if err := c.updateActive(userKey.Key, clusterID); err != nil {
			return errors.HandleCommon(errors.Wrapf(err, "unable to use/activate new api key"), cmd)
		}
	}

	return nil
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	apiKey := args[0]

	userKey, err := c.getAPIKey(apiKey)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	key := &authv1.ApiKey{
		Id:        userKey.Id,
		AccountId: c.config.Auth.Account.Id,
	}

	err = c.client.Delete(context.Background(), key)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return c.config.MaybeDeleteKey(apiKey)
}

func (c *command) use(cmd *cobra.Command, args []string) error {
	apiKey := args[0]

	clusterID, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	err = c.updateActive(apiKey, clusterID)
	if err != nil {
		if !errors.IsUnconfiguredAPIKeyContext(err) {
			return errors.HandleCommon(err, cmd)
		}
		// TODO: interactive prompt for corresponding API key secret
	}
	return nil
}

func (c *command) getAPIKey(key string) (*authv1.ApiKey, error) {
	apiKeys, err := c.client.List(context.Background(), &authv1.ApiKey{AccountId: c.config.Auth.Account.Id})
	if err != nil {
		return nil, err
	}

	var userKey *authv1.ApiKey
	for _, apiKey := range apiKeys {
		if apiKey.Key == key {
			userKey = apiKey
			break
		}
	}

	if userKey == nil {
		return nil, &errors.UnknownAPIKeyError{APIKey: key}
	}

	return userKey, nil
}

func (c *command) createKey(userKey *authv1.ApiKey, environment, clusterID string) error {
	cfg, err := c.config.Context()
	if err != nil {
		return err
	}

	if cfg.KafkaClusters == nil {
		cfg.KafkaClusters = map[string]*config.KafkaClusterConfig{}
	}
	kcc, found := cfg.KafkaClusters[clusterID]
	if !found {
		req := &kafkav1.KafkaCluster{AccountId: environment, Id: clusterID}
		kc, err := c.kafka.Describe(context.Background(), req)
		if err != nil {
			return err
		}

		kcc = &config.KafkaClusterConfig{
			// TODO: registering bootstrap and APIEndpoint is kind of a bad side effect of using an API key
			// These are needed even if no API key is created, such as for managing topics via kafka-api
			Bootstrap:   strings.TrimPrefix(kc.Endpoint, "SASL_SSL://"),
			APIEndpoint: kc.ApiEndpoint,
			APIKey:      userKey.Key,
			APIKeys:     make(map[string]*config.APIKeyPair),
		}

		cfg.KafkaClusters[clusterID] = kcc
	}
	kcc.APIKeys[userKey.Key] = &config.APIKeyPair{
		Key:    userKey.Key,
		Secret: userKey.Secret,
	}
	return c.config.Save()
}

func (c *command) updateActive(apiKey, clusterID string) error {
	cfg, err := c.config.Context()
	if err != nil {
		return err
	}

	_, err = c.getAPIKey(apiKey)
	if err != nil {
		return err
	}

	// TODO: cluster (flag and active) handling will be cleaned up as part of CLI-112
	// In PR: https://github.com/confluentinc/cli/pull/146/files#diff-7bf5d7c832065ed38ccd25c6c525b13bR148
	cluster, found := cfg.KafkaClusters[clusterID]
	if !found {
		return fmt.Errorf("unknown kafka cluster: %s", clusterID)
	}

	_, found = cluster.APIKeys[apiKey]
	if !found {
		// check if this is API key exists server-side
		_, err := c.getAPIKey(apiKey)
		if err != nil {
			return err
		}
		// this means it exists, but we just don't have it saved locally
		return &errors.UnconfiguredAPIKeyContextError{APIKey: apiKey, ClusterID: clusterID}
	}

	cluster.APIKey = apiKey
	return c.config.Save()
}
