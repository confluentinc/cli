package apiKey

import (
	"fmt"
	"context"
	"github.com/codyaray/go-printer"
	"github.com/confluentinc/cli/shared/api-key"

	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	chttp "github.com/confluentinc/ccloud-sdk-go"
)

type command struct {
	*cobra.Command
	config *shared.Config
	client chttp.APIKey
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
	check(createCmd.MarkFlagRequired("id"))
	createCmd.Flags().String("clusterId", "", "cluster id")
	check(createCmd.MarkFlagRequired("clusterId"))
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete API key",
		RunE:  c.delete,
		Args:  cobra.NoArgs,
	}
	deleteCmd.Flags().Int32("id", 0,"service account id")
	check(deleteCmd.MarkFlagRequired("id"))
	deleteCmd.Flags().String("clusterId", "", "cluster id")
	check(deleteCmd.MarkFlagRequired("clusterId"))
	c.AddCommand(deleteCmd)

	return nil
}

func (c *command) create(cmd *cobra.Command, args []string) error {

	id, err := cmd.Flags().GetInt32("id")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	clusterId, err := cmd.Flags().GetString("clusterId")
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
		LogicalClusters: []*authv1.ApiKey_Cluster{
			&authv1.ApiKey_Cluster{Id: clusterId},
	    },
	}

	userKey, errRet := c.client.Create(context.Background(), key)

	if errRet != nil {
		return common.HandleError(errRet, cmd)
	}

	_ = printer.RenderTableOut(userKey, describeFields, describeRenames, os.Stdout)
	fmt.Println("Please Save the API Key and secret.")
	return nil

}

func (c *command) delete(cmd *cobra.Command, args []string) error {

	id, err := cmd.Flags().GetInt32("id")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	clusterId, err := cmd.Flags().GetString("clusterId")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	key := &authv1.ApiKey{
		UserId:      id,
		LogicalClusters: []*authv1.ApiKey_Cluster{
			&authv1.ApiKey_Cluster{Id: clusterId},
		},
	}

	errRet := c.client.Delete(context.Background(), key)

	if errRet != nil {
		return common.HandleError(errRet, cmd)
	}

	return nil

}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
