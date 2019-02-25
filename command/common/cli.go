//go:generate mocker --prefix "" --dst ../../mock/provider_factory.go --pkg mock --selfpkg github.com/confluentinc/cli cli.go ProviderFactory
//go:generate mocker --prefix "" --dst ../../mock/provider.go --pkg mock --selfpkg github.com/confluentinc/cli cli.go Provider

package common

import (
	"fmt"
	"os/exec"
	"reflect"

	"github.com/codyaray/go-editor"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/spf13/cobra"

	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	"github.com/confluentinc/cli/shared"
)

var messages = map[error]string{
	shared.ErrNoContext:      "You must login to access Confluent Cloud.",
	shared.ErrUnauthorized:   "You must login to access Confluent Cloud.",
	shared.ErrExpiredToken:   "Your access to Confluent Cloud has expired. Please login again.",
	shared.ErrIncorrectAuth:  "You have entered an incorrect username or password. Please try again.",
	shared.ErrMalformedToken: "Your auth token has been corrupted. Please login again.",
	shared.ErrNotImplemented: "Sorry, this functionality is not yet available in the CLI.",
	shared.ErrNotFound:       "Kafka cluster not found.", // TODO: parametrize ErrNotFound for better error messaging
}

// HandleError provides standard error messaging for common errors.
func HandleError(err error, cmd *cobra.Command) error {
	// Give an indication of successful completion
	if err == nil {
		return nil
	}

	out := cmd.OutOrStderr()
	if msg, ok := messages[err]; ok {
		fmt.Fprintln(out, msg)
		cmd.SilenceUsage = true
		return err
	}

	switch err.(type) {
	case shared.NotAuthenticatedError:
		fmt.Fprintln(out, err)
	case editor.ErrEditing:
		fmt.Fprintln(out, err)
	case shared.KafkaError:
		fmt.Fprintln(out, err)
	default:
		return err
	}

	return nil
}

type ProviderFactory interface {
	CreateProvider(name string) Provider
}

type ProviderFactoryImpl struct {}

func (f *ProviderFactoryImpl) CreateProvider(name string) Provider {
	return &PluginLoader{Name: name}
}

// Provider loads a plugin
type Provider interface {
	LookupPlugin() (string, error)
	LoadPlugin(interface{}) error
}

// PluginLoader is a helper for finding and instantiating a plugin
type PluginLoader struct {
	Name string
}

// GRPCLaoder returns a new PluginLoader for the given plugin
func GRPCLoader(name string) *PluginLoader {
	return &PluginLoader{Name: name}
}

// LookupPlugin returns the path to a plugin or an error if its not found
func (l *PluginLoader) LookupPlugin() (string, error) {
	runnable, err := exec.LookPath(l.Name)
	if err != nil {
		return "", fmt.Errorf("failed to find plugin: %s", err)
	}
	return runnable, nil
}

// LoadPlugin starts the plugin running as a GRPC server
func (l *PluginLoader) LoadPlugin(value interface{}) error {
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

// Cluster returns the current cluster context
func Cluster(config *shared.Config) (*kafkav1.KafkaCluster, error) {
	ctx, err := config.Context()
	if err != nil {
		return nil, err
	}

	conf, err := config.KafkaClusterConfig()
	if err != nil {
		return nil, err
	}

	return &kafkav1.KafkaCluster{AccountId: config.Auth.Account.Id, Id: ctx.Kafka, ApiEndpoint: conf.APIEndpoint}, nil
}
