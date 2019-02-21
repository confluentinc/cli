package apiKey

import (
	"context"
	"fmt"
	"os"

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
	describeFields  = []string{"Key", "Secret", "UserId"}
	describeRenames = map[string]string{"Key": "API Key", "UserId": "Service Account Id"}
)

// grpcLoader is the default client loader for the CLI
func grpcLoader(i interface{}) error {
	return common.LoadPlugin(apiKey.Name, i)
}

// New returns the Cobra command for API Key.
func New(config *shared.Config) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "api-key",
			Short: "Manage API Key",
		},
		config: config,
	}
	err := cmd.init(grpcLoader)
	return cmd.Command, err
}

func (c *command) init(plugin common.Provider) error {
	c.Command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := c.config.CheckLogin(); err != nil {
			fmt.Printf("failed initial login check \n\n%+v\n", c.config)
			return err
		}
		// Lazy load plugin to avoid unnecessarily spawning child processes
		return plugin(&c.client)
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create API Key.",
		RunE:  c.create,
		Args:  cobra.NoArgs,
	}
	createCmd.Flags().Int32("userid", 0, "service account id")
	_ = createCmd.MarkFlagRequired("userid")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete API key",
		RunE:  c.delete,
		Args:  cobra.NoArgs,
	}
	deleteCmd.Flags().Int32("userid", 0, "service account id")
	_ = deleteCmd.MarkFlagRequired("userid")
	deleteCmd.Flags().String("apikey", "", "api Key")
	_ = deleteCmd.MarkFlagRequired("apikey")
	c.AddCommand(deleteCmd)

	return nil
}

func (c *command) create(cmd *cobra.Command, args []string) error {

	userId, err := cmd.Flags().GetInt32("userid")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	cluster, err := common.Cluster(c.config)
	if err != nil {
		return common.HandleError(err, cmd)
	}
	description := "Service Account API Key"

	key := &authv1.ApiKey{
		UserId:      userId,
		Description: description,
		AccountId: c.config.Auth.Account.Id,
		LogicalClusters: []*authv1.ApiKey_Cluster{
			{Id: cluster.Id},
		},
	}

	userKey, errRet := c.client.Create(context.Background(), key)

	if errRet != nil {
		return common.HandleError(errRet, cmd)
	}

	fmt.Println("Please Save the API Key ID, API Key and Secret.")
	var stdout = os.Stdout
	return printer.RenderTableOut(userKey, describeFields, describeRenames, stdout)
}

func getApiKeyId(apiKeys []*authv1.ApiKey, apiKey string)(int32, error) {
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

	userId, err := cmd.Flags().GetInt32("userid")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	apiKeys, err := c.client.List(context.Background(), &authv1.ApiKey{AccountId: c.config.Auth.Account.Id});
	if err != nil {
		return common.HandleError(err, cmd)
	}

	id, err := getApiKeyId(apiKeys, apiKey)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	key := &authv1.ApiKey{
		Id: id,
		UserId: userId,
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

