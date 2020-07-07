package config

// Example command created by @nkuo to describe the parts to adding a command

import (
	"github.com/spf13/cobra"
	"strconv"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

// the internals that this command needs (passed in from initialization of ccloud/confluent object) + cobra.Command & testing internals?
type fileCommand struct {
	*pcmd.CLICommand                // seems to be a wrapper for *cobra.Command, so replaces it
	prerunner        pcmd.PreRunner // every up-to-date command seems to have this passed in to use pcmd, manages authentication stuff?

	// any other things that this command needs
	config *v3.Config
}

// this is called by the resource's command.go, and it is registered by it
func NewFile(prerunner pcmd.PreRunner, config *v3.Config) *cobra.Command {
	// Create the pcmd CLI Command, which is a wrapper for cobra.Command
	cliCmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "file",
			Short: "Manage the config file.",
		}, prerunner)

	// Create fileCommand, wrapper defined above^
	cmd := &fileCommand{
		CLICommand: cliCmd,
		prerunner:  prerunner,
		config:     config,
	}
	// Call init(), CLI codebase convention for registering verbs
	cmd.init()
	return cmd.Command // <- there's some funky stuff going on here with the anonymous type
}

// Register each of the verbs and expected args
func (c *fileCommand) init() {
	// Register the show command
	showCmd := &cobra.Command{
		Use:   "show <num-times>",
		Short: "Show the config file location a specified number of times.",
		RunE:  c.show,             // register the function to call
		Args:  cobra.ExactArgs(1), // specify arguments
	}
	c.AddCommand(showCmd)
}

// the actual function called when config file show <arg> is called
func (c *fileCommand) show(cmd *cobra.Command, args []string) error {
	// cmd specifies ___ and args are command line args
	numTimes, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	filename := c.config.Filename
	if filename == "" {
		return errors.New("No config file exists!")
	}
	for i := 0; i < numTimes; i++ {
		pcmd.Println(cmd, filename)
	}
	return nil
}
