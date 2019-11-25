package context

import (
	"github.com/spf13/cobra"
)

type command struct {
	*cobra.Command
}

// NewContext returns the Cobra contextCommand for `config context`.
func NewContextWIP() *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "context",
			Short: "Manage config contexts.",
		},
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	createCmd := &cobra.Command{
		Use:   "create <context-name>",
		Short: "Create a context",
		Args:  cobra.ExactArgs(1),
		Run: hi,
	}
	createCmd.Flags().String("environment", "", "Sets the environment.")
	createCmd.Flags().String("cluster", "", "Sets the cluster.")
	createCmd.Flags().String("topic", "", "Sets the topic.")
	createCmd.Flags().Bool("kafka-auth", false, "Initialize with a bootstrap url, API key, and API secret. "+
		"Can be done interactively, with flags, or both.")
	createCmd.Flags().String("bootstrap", "", "Bootstrap URL to use with --kafka-auth.")
	createCmd.Flags().String("api-key", "", "API key to use with --kafka-auth.")
	createCmd.Flags().String("api-secret", "", "API secret to use with --kafka-auth. Can be specified as plain text, "+
		"as a file, starting with '@', or as stdin, starting with '-'.")
	createCmd.Flags().String("login-auth", "", "Initialize with a username and password.")
	createCmd.Flags().String("username", "", "Username to use with --login-auth.")
	createCmd.Flags().String("password", "", "Password to use with --login-auth.")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)
}

func hi(cmd *cobra.Command, args []string) {
	
}
