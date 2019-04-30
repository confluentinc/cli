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
)

type command struct {
	*cobra.Command
	config *config.Config
	client ccloud.APIKey
}

var (
	listFields    = []string{"Key", "UserId"}
	listLabels    = []string{"Key", "Owner"}
	createFields  = []string{"Key", "Secret"}
	createRenames = map[string]string{"Key": "API Key"}
)

// New returns the Cobra command for API Key.
func New(prerunner pcmd.PreRunner, config *config.Config, client ccloud.APIKey) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:               "api-key",
			Short:             "Manage API keys",
			PersistentPreRunE: prerunner.Authenticated(),
		},
		config: config,
		client: client,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
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
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	c.AddCommand(&cobra.Command{
		Use:   "delete KEY",
		Short: "Delete API key",
		RunE:  c.delete,
		Args:  cobra.ExactArgs(1),
	})
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
	return printer.RenderTableOut(userKey, createFields, createRenames, os.Stdout)
}

func getApiKeyId(apiKeys []*authv1.ApiKey, apiKey string) (int32, error) {
	var id int32
	for _, key := range apiKeys {
		if key.Key == apiKey {
			id = key.Id
			break
		}
	}

	if id == 0 {
		return id, fmt.Errorf(" Invalid Key")
	}

	return id, nil
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	apiKey := args[0]

	apiKeys, err := c.client.List(context.Background(), &authv1.ApiKey{AccountId: c.config.Auth.Account.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	id, err := getApiKeyId(apiKeys, apiKey)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	key := &authv1.ApiKey{
		Id:        id,
		AccountId: c.config.Auth.Account.Id,
	}

	err = c.client.Delete(context.Background(), key)

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	c.config.MaybeDeleteKey(apiKey)
	return nil
}
