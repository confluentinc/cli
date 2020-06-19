package local

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

func NewDemoCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	demoCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "demo",
			Short: "Run demos provided in GitHub repo https://github.com/confluentinc/examples.",
			Args:  cobra.NoArgs,
		},
		cfg, prerunner)

	demoCommand.AddCommand(NewDemoInfoCommand(prerunner, cfg))
	demoCommand.AddCommand(NewDemoListCommand(prerunner, cfg))
	demoCommand.AddCommand(NewDemoStartCommand(prerunner, cfg))
	demoCommand.AddCommand(NewDemoStopCommand(prerunner, cfg))
	demoCommand.AddCommand(NewDemoUpdateCommand(prerunner, cfg))

	return demoCommand.Command
}

func NewDemoInfoCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	demoInfoCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "info [demo]",
			Short: "Show README for the specified demo.",
			Args:  cobra.ExactArgs(1),
			RunE:  runDemoInfoCommand,
		},
		cfg, prerunner)

	return demoInfoCommand.Command
}

func runDemoInfoCommand(command *cobra.Command, _ []string) error {
	return nil
}

func NewDemoListCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	demoListCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "list",
			Short: "List names of available demos.",
			Args:  cobra.NoArgs,
			RunE:  runDemoListCommand,
		},
		cfg, prerunner)

	return demoListCommand.Command
}

func runDemoListCommand(command *cobra.Command, _ []string) error {
	return nil
}

func NewDemoStartCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	demoStartCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "start [demo]",
			Short: "Start the specified demo.",
			Args:  cobra.ExactArgs(1),
			RunE:  runDemoStartCommand,
		},
		cfg, prerunner)

	return demoStartCommand.Command
}

func runDemoStartCommand(command *cobra.Command, _ []string) error {
	return nil
}

func NewDemoStopCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	demoStopCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "stop [demo]",
			Short: "Stop the specified demo.",
			Args:  cobra.ExactArgs(1),
			RunE:  runDemoStopCommand,
		},
		cfg, prerunner)

	return demoStopCommand.Command
}

func runDemoStopCommand(command *cobra.Command, _ []string) error {
	return nil
}

func NewDemoUpdateCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	demoUpdateCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "update [demo]",
			Short: "Update the specified demo.",
			Args:  cobra.ExactArgs(1),
			RunE:  runDemoUpdateCommand,
		},
		cfg, prerunner)

	return demoUpdateCommand.Command
}

func runDemoUpdateCommand(command *cobra.Command, _ []string) error {
	return nil
}
