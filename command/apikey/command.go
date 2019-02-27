package apikey

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	ccloud "github.com/confluentinc/ccloud-sdk-go"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/apikey"
	"github.com/confluentinc/go-printer"
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

// New returns the Cobra command for API Key.
func New(config *shared.Config, factory common.GRPCPluginFactory) (*cobra.Command, error) {
	return newCMD(config, factory.Create(apikey.Name))
}

// newCMD returns a command for interacting with API Key.
func newCMD(config *shared.Config, provider common.GRPCPlugin) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "api-key",
			Short: "Manage API keys",
		},
		config: config,
	}
	_, err := provider.LookupPath()
	if err != nil {
		return nil, err
	}
	err = cmd.init(provider)
	return cmd.Command, err
}

func (c *command) init(plugin common.GRPCPlugin) error {
	c.Command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := common.SetLoggingVerbosity(cmd, c.config.Logger); err != nil {
			return common.HandleError(err, cmd)
		}
		if err := c.config.CheckLogin(); err != nil {
			return common.HandleError(err, cmd)
		}
		// Lazy load plugin to avoid unnecessarily spawning child processes
		return plugin.Load(&c.client, c.config.Logger)
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
	createCmd.Flags().String("cluster", "", "Grant access to a cluster with this ID")
	_ = createCmd.MarkFlagRequired("cluster")
	createCmd.Flags().Int32("service-account-id", 0, "Create API key for a service account")
	createCmd.Flags().String("description", "", "Description or purpose for the API key")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete API key",
		RunE:  c.delete,
		Args:  cobra.NoArgs,
	}
	deleteCmd.Flags().String("api-key", "", "API key")
	_ = deleteCmd.MarkFlagRequired("api-key")
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
	clusterID, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	userId, err := cmd.Flags().GetInt32("service-account-id")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	key := &authv1.ApiKey{
		UserId:          userId,
		Description:     description,
		AccountId:       c.config.Auth.Account.Id,
		LogicalClusters: []*authv1.ApiKey_Cluster{{Id: clusterID}},
	}

	userKey, err := c.client.Create(context.Background(), key)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	fmt.Println("Please save the API Key and Secret. THIS IS THE ONLY CHANCE YOU HAVE!")
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
	apiKey, err := cmd.Flags().GetString("api-key")
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
		AccountId: c.config.Auth.Account.Id,
	}

	err = c.client.Delete(context.Background(), key)

	if err != nil {
		return common.HandleError(err, cmd)
	}

	return nil
}
