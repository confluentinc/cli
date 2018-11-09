package common

import (
	"fmt"
	"os/exec"
	"reflect"

	"github.com/codyaray/go-editor"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/spf13/cobra"

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
	out := cmd.OutOrStderr()
	if msg, ok := messages[err]; ok {
		fmt.Fprintln(out, msg)
		return err
	}

	switch err.(type) {
	case editor.ErrEditing:
		fmt.Fprintln(out, err)
	case shared.NotAuthenticatedError:
		fmt.Fprintln(out, err)
	case shared.KafkaError:
		fmt.Fprintln(out, err)
	default:
		return err
	}
	return nil
}

// LoadPlugin starts a GRPC server identified by name
func LoadPlugin(name string, value interface{}) error {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("value of type %T must be a pointer for a GRPC client", value)
	}

	runnable, err := exec.LookPath(name)
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
			Level:  hclog.Info,
			Name:   "plugin",
		}),
	})

	// Connect via RPC.
	rpcClient, err := client.Client()
	if err != nil {
		return err
	}

	// Request the plugin
	obj2, err := rpcClient.Dispense(name)
	if err != nil {
		return err
	}
	rv.Elem().Set(reflect.ValueOf(reflect.ValueOf(obj2).Interface()))
	return err
}

// IsSet returns true if an interface has a non-zero value set
func IsSet(v interface{}) bool {
	return !reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}
