package apiKey

import (
	"context"
	"fmt"
	"github.com/codyaray/go-printer"
	"github.com/confluentinc/cli/shared/api-key"

	"os"

	"github.com/spf13/cobra"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
)

type command struct {
	*cobra.Command
	config *shared.Config
	client chttp.APIKey
}

var (
	describeFields  = []string{"Id", "Key", "Secret", "UserId"}
	describeRenames = map[string]string{"Id": "Api Key Id","Key": "Api Key", "UserId": "Service Account Id"}
)

// grpcLoader is the default client loader for the CLI
func grpcLoader(i interface{}) error {
	return common.LoadPlugin(apiKey.Name, i)
}

// New returns the Cobra command for API Key.
func New(config *shared.Config) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "api-keys",
			Short: "Manage API Keys",
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
	createCmd.Flags().Int32("id", 0, "service account id")
	createCmd.MarkFlagRequired("id")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete API key",
		RunE:  c.delete,
		Args:  cobra.NoArgs,
	}
	deleteCmd.Flags().Int32("userId", 0, "service account id")
	deleteCmd.MarkFlagRequired("userId")
	deleteCmd.Flags().Int32("apiKeyId", 0, "api Key id")
	deleteCmd.MarkFlagRequired("apiKeyId")
	c.AddCommand(deleteCmd)

	return nil
}

func (c *command) create(cmd *cobra.Command, args []string) error {

	id, err := cmd.Flags().GetInt32("id")
	fmt.Println("User ID: ", id)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	cluster, err := common.Cluster(c.config)
	if err != nil {
		return common.HandleError(err, cmd)
	}
	description := "Service Account API Key"
	if err != nil {
		return common.HandleError(err, cmd)
	}

	key := &authv1.ApiKey{
		UserId:      id,
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

	fmt.Println("Please Save the API Key and secret.")
	return printer.RenderTableOut(userKey, describeFields, describeRenames, os.Stdout)
}

func (c *command) delete(cmd *cobra.Command, args []string) error {

	id, err := cmd.Flags().GetInt32("apiKeyId")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	cluster, err := common.Cluster(c.config)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	userId, err := cmd.Flags().GetInt32("userId")
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

	errRet := c.client.Delete(context.Background(), key)

	if errRet != nil {
		return common.HandleError(errRet, cmd)
	}

	return nil

}

