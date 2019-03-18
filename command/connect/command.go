package connect

/*import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/connect"
)

type command struct {
	*cobra.Command
	config *shared.Config
}

// New returns the default command object for interacting with Connect.
func New(config *shared.Config, factory common.GRPCPluginFactory) (*cobra.Command, error) {
	return newCMD(config, factory.Create(connect.Name))
}

// NewConnectCommand returns a command object using a custom Connect provider.
func NewConnectCommand(config *shared.Config, provider common.GRPCPlugin) (*cobra.Command, error) {
	return newCMD(config, provider)
}

// newCMD returns a command for interacting with Connect.
func newCMD(config *shared.Config, provider common.GRPCPlugin) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "connect",
			Short: "Manage Kafka Connect",
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
	sinkCmd, err := NewSink(c.config, plugin)
	if err != nil {
		return err
	}
	c.AddCommand(sinkCmd)

	return nil
}*/
