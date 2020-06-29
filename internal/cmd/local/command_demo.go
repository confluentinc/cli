package local

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/local"
)

var supportedDemos = []string{
	"ccloud",
	"connect-streams-pipeline",
	"cp-quickstart",
	"microservices-orders",
	"music",
}

func NewDemoCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "demo",
			Short: "Run demos provided at https://github.com/confluentinc/examples.",
			Args:  cobra.NoArgs,
		}, prerunner)

	c.AddCommand(NewDemoInfoCommand(prerunner))
	c.AddCommand(NewDemoListCommand(prerunner))
	c.AddCommand(NewDemoStartCommand(prerunner))
	c.AddCommand(NewDemoStopCommand(prerunner))

	// TODO: Show once demos are updated with new confluent local syntax
	c.Hidden = true

	return c.Command
}

func NewDemoInfoCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "info [demo]",
			Short: "Show the README for a demo.",
			Args:  cobra.ExactArgs(1),
		}, prerunner)

	c.Command.RunE = c.runDemoInfoCommand
	return c.Command
}

func (c *LocalCommand) runDemoInfoCommand(command *cobra.Command, args []string) error {
	demo := args[0]
	if !local.Contains(supportedDemos, demo) {
		return fmt.Errorf("demo not supported: %s", demo)
	}

	if err := c.fetchExamplesRepo(); err != nil {
		return err
	}

	readme, err := c.ch.ReadDemoReadme(demo)
	if err != nil {
		return err
	}

	command.Println(readme)
	return nil
}

func NewDemoListCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List available demos.",
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.Run = c.runDemoListCommand
	return c.Command
}

func (c *LocalCommand) runDemoListCommand(command *cobra.Command, _ []string) {
	list := local.BuildTabbedList(supportedDemos)
	command.Println("Available demos:")
	command.Println(list)
	command.Println("To start a demo, run 'confluent local demo start [demo]'")
}

func NewDemoStartCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "start [demo]",
			Short: "Start a demo.",
			Args:  cobra.ExactArgs(1),
		}, prerunner)

	c.Command.RunE = c.runDemoStartCommand
	return c.Command
}

func (c *LocalCommand) runDemoStartCommand(_ *cobra.Command, args []string) error {
	return c.run(args[0], "start.sh")
}

func NewDemoStopCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "stop [demo]",
			Short: "Stop a demo.",
			Args:  cobra.ExactArgs(1),
		}, prerunner)

	c.Command.RunE = c.runDemoStopCommand
	return c.Command
}

func (c *LocalCommand) runDemoStopCommand(_ *cobra.Command, args []string) error {
	return c.run(args[0], "stop.sh")
}

func (c *LocalCommand) fetchExamplesRepo() error {
	hasRepo, err := c.ch.HasFile("examples")
	if err != nil {
		return err
	}

	dir, err := c.ch.GetFile("examples")
	if err != nil {
		return err
	}

	var repo *git.Repository

	if hasRepo {
		repo, err = git.PlainOpen(dir)
		if err != nil {
			return err
		}
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

	if hasRepo {
		err := tree.Pull(&git.PullOptions{})
		if err != nil && err.Error() != "already up-to-date" {
			return err
		}
	}

	version, err := c.ch.GetConfluentVersion()
	if err != nil {
		return err
	}

	branch := plumbing.NewBranchReferenceName(fmt.Sprintf("%s-post", version))
	if err := tree.Checkout(&git.CheckoutOptions{Branch: branch}); err != nil {
		return err
	}

	return nil
}

func (c *LocalCommand) run(demo string, script string) error {
	if !local.Contains(supportedDemos, demo) {
		return fmt.Errorf("demo not supported: %s", demo)
	}

	if err := c.fetchExamplesRepo(); err != nil {
		return err
	}

	scriptFile, err := c.ch.GetFile("examples", demo, script)
	if err != nil {
		return err
	}

	dir, err := c.ch.GetFile("examples", demo)
	if err != nil {
		return err
	}

	command := exec.Command(scriptFile)
	command.Dir = dir
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	return command.Run()
}
