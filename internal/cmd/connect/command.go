package connect

/*import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
)

type command struct {
	*cobra.Command
	config *shared.Config
	client ccloud.Connect
}

// New returns the default command object for interacting with Connect.
func New(config *shared.Config, client ccloud.Connect) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "connect",
			Short: "Manage Kafka Connect.",
		},
		config: config,
		client: client,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	sinkCmd, err := NewSink(c.config, c.client)
	if err != nil {
		return err
	}
	c.AddCommand(sinkCmd)
}*/
