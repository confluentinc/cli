package local

import (
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/local"
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
	ch := local.NewConfluentHomeManager()

	hasRepo, err := ch.HasExamplesRepo()
	if err != nil {
		return err
	}

	dir, err := ch.GetExamplesRepo()
	if err != nil {
		return err
	}

	var repo *git.Repository

	if hasRepo {
		repo, err = git.PlainOpen(dir)
		if err != nil {
			return err
		}
		// TODO: Update
	} else {
		repo, err = git.PlainClone(dir, false, &git.CloneOptions{
			URL: "https://github.com/confluentinc/examples.git",
		})
		if err != nil {
			return err
		}
	}

	tree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = tree.Checkout(&git.CheckoutOptions{
		Branch: "5.5.0-post",
	})
	if err != nil {
		return err
	}

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
