//go:generate mocker --prefix "" --dst ../../mock/plugin_factory.go --pkg mock --selfpkg github.com/confluentinc/cli plugin.go GRPCPluginFactory
//go:generate mocker --prefix "" --dst ../../mock/plugin.go --pkg mock --selfpkg github.com/confluentinc/cli plugin.go GRPCPlugin

package common

import (
	"fmt"
	"os/exec"
	"reflect"

	"github.com/confluentinc/cli/shared"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

// GRPCPluginFactory creates GRPCServices
type GRPCPluginFactory interface {
	Create(name string) GRPCPlugin
}

// GRPCPluginFactoryImpl loads GRPCPlugins from the PATH
type GRPCPluginFactoryImpl struct{}

// Create returns the default GRPCPlugin for the named plugin
func (f *GRPCPluginFactoryImpl) Create(name string) GRPCPlugin {
	return &GRPCPluginImpl{Name: name}
}

// GRPCPlugin represents a plugin that can be found and loaded on the CLI
type GRPCPlugin interface {
	LookupPath() (string, error)
	Load(interface{}) error
}

// GRPCPluginImpl finds and instantiates plugins on the PATH
type GRPCPluginImpl struct {
	Name string
}

// LookupPath returns the path to a plugin or an error if its not found
func (l *GRPCPluginImpl) LookupPath() (string, error) {
	runnable, err := exec.LookPath(l.Name)
	if err != nil {
		return "", fmt.Errorf("failed to find plugin: %s", err)
	}
	return runnable, nil
}

// Load starts the plugin running as a GRPC server and sets the client in value
func (l *GRPCPluginImpl) Load(value interface{}) error {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("value of type %T must be a pointer for a GRPC client", value)
	}

	runnable, err := exec.LookPath(l.Name)
	if err != nil {
		return fmt.Errorf("failed to load plugin: %s", err)
	}

	// We're a host. Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  shared.Handshake,
		Plugins:          shared.PluginMap,
		Cmd:              exec.Command("sh", "-c", runnable), // nolint: gas
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Managed:          true,
		Logger: hclog.New(&hclog.LoggerOptions{
			Output: hclog.DefaultOutput,
			Level:  hclog.Error,
			Name:   "plugin",
		}),
	})

	// Connect via RPC.
	rpcClient, err := client.Client()
	if err != nil {
		return err
	}

	// Request the plugin
	impl, err := rpcClient.Dispense(l.Name)
	if err != nil {
		return err
	}
	rv.Elem().Set(reflect.ValueOf(reflect.ValueOf(impl).Interface()))
	return err
}
