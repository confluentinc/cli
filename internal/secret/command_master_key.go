package secret

import (
	"github.com/spf13/cobra"
)

func (c *command) newMasterKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "master-key",
		Short: "Manage the master key for Confluent Platform.",
	}

	cmd.AddCommand(c.newGenerateFunction())

	return cmd
}
