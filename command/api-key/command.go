package apiKey

import (
	"fmt"
	"context"
	"github.com/codyaray/go-printer"

	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
)

type command struct {
	*cobra.Command
	config *shared.Config
	key  ApiKey
}

var (
	describeFields  = []string{"Key", "Secret", "UserId"}
	describeRenames = map[string]string{"Key": "API Key", "UserId": "Service Account Id"}
)

// New returns the Cobra command for API Key.
func New(config *shared.Config) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "api-keys",
			Short: "Manage API Keys",
		},
		config: config,
	}
	err := cmd.init()
	return cmd.Command, err
}

func (c *command) init() error {
	path, err := exec.LookPath("confluent-api-key-plugin")
	if err != nil {
		return fmt.Errorf("skipping api-key: plugin isn't installed")
	}

	// We're a host. Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  shared.Handshake,
		Plugins:          shared.PluginMap,
		Cmd:              exec.Command("sh", "-c", path), // nolint: gas
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Managed:          true,
		Logger: hclog.New(&hclog.LoggerOptions{
			Output: hclog.DefaultOutput,
			Level:  hclog.Info,
			Name:   "plugin",
		}),
	})

	// Connect via RPC.
	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("apiKey")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Got a client now communicating over RPC.
	c.key = raw.(ApiKey)

	// All commands require login first
	c.Command.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if err = c.config.CheckLogin(); err != nil {
			_ = common.HandleError(err, cmd)
			os.Exit(1)
		}
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


	key := &schedv1.ApiKey{
		UserId:      id,
		Description: description,
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			&schedv1.ApiKey_Cluster{Id: clusterId},
	    },
	}

	userKey, errRet := c.key.Create(context.Background(), key)

	if errRet != nil {
		return common.HandleError(errRet, cmd)
	}

	printer.RenderTableOut(userKey, describeFields, describeRenames, os.Stdout)
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

	key := &schedv1.ApiKey{
		UserId:      id,
		LogicalClusters: []*schedv1.ApiKey_Cluster{
			&schedv1.ApiKey_Cluster{Id: clusterId},
		},
	}

	errRet := c.key.Delete(context.Background(), key)

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
