package local

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

const longDescription = `Use these commands to try out Confluent Platform by running a single-node
instance locally on your machine. This set of commands are NOT intended for production use.

You must download and install Confluent Platform from https://www.confluent.io/download on your
machine. These commands require the path to the installation directory via the --path flag or
the CONFLUENT_HOME environment variable.

You can use these commands to explore, test, experiment, and otherwise familiarize yourself
with Confluent Platform.

DO NOT use these commands to setup or manage Confluent Platform in production.
`

type command struct {
	*cobra.Command
	shell ShellRunner
}

// New returns the Cobra command for `local`.
func New(prerunner pcmd.PreRunner, shell ShellRunner) *cobra.Command {
	localCmd := &command{
		Command: &cobra.Command{
			Use:               "local",
			Short:             "Manage a local Confluent Platform development environment.",
			Long:              longDescription,
			Args:              cobra.ArbitraryArgs,
			PersistentPreRunE: prerunner.Anonymous(),
		},
		shell: shell,
	}
	localCmd.Command.RunE = localCmd.run
	localCmd.Flags().String("path", "", "Path to Confluent Platform install directory.")
	localCmd.Flags().SortFlags = false
	// This is used for "confluent help local foo" and "confluent local foo --help"
	localCmd.Command.SetHelpFunc(localCmd.help)
	return localCmd.Command
}

func (c *command) run(cmd *cobra.Command, args []string) error {
	path, err := cmd.Flags().GetString("path")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if path == "" {
		if home, found := os.LookupEnv("CONFLUENT_HOME"); found {
			path = home
		} else if len(args) != 0 { // if no args specified, allow so we just show usage
			return fmt.Errorf("Pass --path /path/to/confluent flag or set environment variable CONFLUENT_HOME")
		}
	}
	err = c.runBashCommand(path, "main", args)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return nil
}

func (c *command) help(cmd *cobra.Command, args []string) {
	// if "confluent help local foo bar" is called, args is empty, so we just show usage :(
	// if "confluent local foo bar --help" is called, args is [local, foo, bar, --help]
	// transform args: drop first "local" and any "--help" flag. [local, foo, bar, --help] -> [help, foo, bar]
	if len(args) > 0 && args[0] == "local" {
		args = args[1:]
	}
	var a []string
	for _, arg := range args {
		if arg != "--help" {
			a = append(a, arg)
		}
	}
	_ = c.runBashCommand("", "help", a)
}

func (c *command) runBashCommand(path string, command string, args []string) error {
	c.shell.Init(os.Stdout, os.Stderr)
	c.shell.Export("CONFLUENT_HOME", path)
	err := c.shell.Source("cp_cli/confluent.sh", Asset)
	if err != nil {
		return err
	}

	_, err = c.shell.Run(command, args)
	if err != nil {
		return err
	}
	return nil
}
