package api_key

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/codyaray/go-printer"
	"github.com/spf13/cobra"

	ccloud "github.com/confluentinc/ccloud-sdk-go"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/api-key"
)

type command struct {
	*cobra.Command
	config *shared.Config
	client ccloud.APIKey
}

var (
	listFields    = []string{"Key", "UserId", "LogicalClusters"}
	listLabels    = []string{"Key", "Owner", "Clusters"}
	createFields  = []string{"Key", "Secret"}
	createRenames = map[string]string{"Key": "API Key"}
)

// grpcLoader is the default client loader for the CLI
func grpcLoader(i interface{}) error {
	return common.LoadPlugin(api_key.Name, i)
}

// New returns the Cobra command for API Key.
func New(config *shared.Config) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "api-key",
			Short: "Manage API keys",
		},
		config: config,
	}
	err := cmd.init(grpcLoader)
	return cmd.Command, err
}

func (c *command) init(plugin common.Provider) error {
	c.Command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := c.config.CheckLogin(); err != nil {
			return common.HandleError(err, cmd)
		}
		// Lazy load plugin to avoid unnecessarily spawning child processes
		return plugin(&c.client)
	}

	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List API keys",
		RunE:  c.list,
	})

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create API key",
		RunE:  c.create,
		Args:  cobra.NoArgs,
	}
	createCmd.Flags().Int32("serviceaccountid", 0, "service account id")
	createCmd.Flags().String("description", "", "description for api key")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete API key",
		RunE:  c.delete,
		Args:  cobra.NoArgs,
	}
	deleteCmd.Flags().Int32("serviceaccountid", 0, "service account id")
	deleteCmd.Flags().String("apikey", "", "api Key")
	_ = deleteCmd.MarkFlagRequired("apikey")
	c.AddCommand(deleteCmd)

	return nil
}

func (c *command) list(cmd *cobra.Command, args []string) error {
	apiKeys, err := c.client.List(context.Background(), &authv1.ApiKey{AccountId: c.config.Auth.Account.Id})
	if err != nil {
		return common.HandleError(err, cmd)
	}

	type keyDisplay struct {
		Key             string
		Description     string
		UserId          int32
		LogicalClusters string
	}

	var data [][]string
	for _, apiKey := range apiKeys {
		// ignore keys owned by Confluent-internal user (healthcheck, etc)
		if apiKey.UserId == 0 {
			continue
		}
		var clusters []string
		for _, c := range apiKey.LogicalClusters {
			buf := new(bytes.Buffer)
			buf.WriteString(c.Id)
			// TODO: uncomment once we migrate DB so all API keys have a type
			//buf.WriteString(" (type=")
			//buf.WriteString(c.Type)
			//buf.WriteString(")")
			clusters = append(clusters, buf.String())
		}
		data = append(data, printer.ToRow(&keyDisplay{
			Key:             apiKey.Key,
			Description:     apiKey.Description,
			UserId:          apiKey.UserId,
			LogicalClusters: strings.Join(clusters, ", "),
		}, listFields))
	}

	printer.RenderCollectionTable(data, listLabels)
	return nil
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	userId, err := cmd.Flags().GetInt32("serviceaccountid")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	cluster, err := common.Cluster(c.config)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	key := &authv1.ApiKey{
		UserId:      userId,
		Description: description,
		AccountId:   c.config.Auth.Account.Id,
		LogicalClusters: []*authv1.ApiKey_Cluster{
			{Id: cluster.Id},
		},
	}

	userKey, err := c.client.Create(context.Background(), key)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	fmt.Println("Please Save the API Key ID, API Key and Secret.")
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
	apiKey, err := cmd.Flags().GetString("apikey")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	cluster, err := common.Cluster(c.config)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	userId, err := cmd.Flags().GetInt32("serviceaccountid")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	apiKeys, err := c.client.List(context.Background(), &authv1.ApiKey{AccountId: c.config.Auth.Account.Id})
	if err != nil {
		return common.HandleError(err, cmd)
	}

	id, err := getApiKeyId(apiKeys, apiKey)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	key := &authv1.ApiKey{
		Id:        id,
		UserId:    userId,
		AccountId: c.config.Auth.Account.Id,
		LogicalClusters: []*authv1.ApiKey_Cluster{
			{Id: cluster.Id},
		},
	}

	err = c.client.Delete(context.Background(), key)

	if err != nil {
		return common.HandleError(err, cmd)
	}

	return nil
}
