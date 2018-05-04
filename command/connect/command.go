package connect

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/spf13/cobra"

	common "github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
)

type Command struct {
	*cobra.Command
	*shared.Config
	connect Connect
}

func New(config *shared.Config) (*cobra.Command, error) {
	cmd := &Command{
		Command: &cobra.Command{
			Use:   "connect",
			Short: "Manage connectors.",
		},
		Config: config,
	}
	err := cmd.init()
	return cmd.Command, err
}

func (c *Command) init() error {
	path, err := exec.LookPath("confluent-connect-plugin")
	if err != nil {
		return fmt.Errorf("skipping connect: plugin isn't installed")
	}

	// We're a host. Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  shared.Handshake,
		Plugins:          shared.PluginMap,
		Cmd:              exec.Command("sh", "-c", path),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	// Connect via RPC.
	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("connect")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Got a client now communicating over RPC.
	c.connect = raw.(Connect)

	c.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Create a connector.",
		RunE:  c.create,
	})
	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List connectors.",
		RunE:  c.list,
	})
	c.AddCommand(&cobra.Command{
		Use:   "describe",
		Short: "Describe a connector.",
		RunE:  c.describe,
	})
	c.AddCommand(&cobra.Command{
		Use:   "delete",
		Short: "Delete a connector.",
		RunE:  c.delete,
	})
	c.AddCommand(&cobra.Command{
		Use:   "update",
		Short: "Update a connector.",
		RunE:  c.update,
	})
	c.AddCommand(&cobra.Command{
		Use:   "auth",
		Short: "Auth a connector.",
		RunE:  c.auth,
	})

	return nil
}

func (c *Command) list(Command *cobra.Command, args []string) error {
	connectors, err := c.connect.List(context.Background())
	if err != nil {
		return common.HandleError(err)
	}
	fmt.Println(connectors)
	return nil
}

func (c *Command) create(Command *cobra.Command, args []string) error {
	return shared.ErrNotImplemented
}

func (c *Command) describe(Command *cobra.Command, args []string) error {
	return shared.ErrNotImplemented
}

func (c *Command) update(Command *cobra.Command, args []string) error {
	return shared.ErrNotImplemented
}

func (c *Command) delete(Command *cobra.Command, args []string) error {
	return shared.ErrNotImplemented
}

func (c *Command) auth(Command *cobra.Command, args []string) error {
	return shared.ErrNotImplemented
}
