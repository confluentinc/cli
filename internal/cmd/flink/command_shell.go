package flink

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (c *command) newStartFlinkSqlClientCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Start Flink interactive shell.",
		RunE:  c.startFlinkSqlClient,
	}

	return cmd
}

func (c *command) startFlinkSqlClient(cmd *cobra.Command, args []string) error {
	fmt.Println("hello world")

	return nil
}
