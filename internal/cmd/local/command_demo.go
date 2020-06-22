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
	demoCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "demo",
			Short: "Run demos provided at https://github.com/confluentinc/examples.",
			Args:  cobra.NoArgs,
		}, prerunner)

	demoCommand.AddCommand(NewDemoInfoCommand(prerunner))
	demoCommand.AddCommand(NewDemoListCommand(prerunner))
	demoCommand.AddCommand(NewDemoStartCommand(prerunner))
	demoCommand.AddCommand(NewDemoStopCommand(prerunner))

	demoCommand.Hidden = true

	return demoCommand.Command
}

func NewDemoInfoCommand(prerunner cmd.PreRunner) *cobra.Command {
	demoInfoCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "info [demo]",
			Short: "Show the README for a demo.",
			Args:  cobra.ExactArgs(1),
			RunE:  runDemoInfoCommand,
		}, prerunner)

	return demoInfoCommand.Command
}

func runDemoInfoCommand(command *cobra.Command, args []string) error {
	demo := args[0]
	if !isSupported(demo) {
		return fmt.Errorf("demo not supported: %s", demo)
	}

	ch := local.NewConfluentHomeManager()

	if err := fetchExamplesRepo(ch); err != nil {
		return err
	}

	readme, err := ch.GetDemoReadme(demo)
	if err != nil {
		return err
	}

	command.Println(readme)
	return nil
}

func NewDemoListCommand(prerunner cmd.PreRunner) *cobra.Command {
	demoListCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "list",
			Short: "List available demos.",
			Args:  cobra.NoArgs,
			Run:   runDemoListCommand,
		}, prerunner)

	return demoListCommand.Command
}

func runDemoListCommand(command *cobra.Command, _ []string) {
	list := local.BuildTabbedList(supportedDemos)
	command.Println("Available demos:")
	command.Println(list)
	command.Println("To start a demo, run 'confluent local demo start [demo]'")
}

func NewDemoStartCommand(prerunner cmd.PreRunner) *cobra.Command {
	demoStartCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "start [demo]",
			Short: "Start a demo.",
			Args:  cobra.ExactArgs(1),
			RunE:  runDemoStartCommand,
		}, prerunner)

	return demoStartCommand.Command
}

func runDemoStartCommand(command *cobra.Command, args []string) error {
	ch := local.NewConfluentHomeManager()
	return run(ch, args[0], "start.sh")
}

func NewDemoStopCommand(prerunner cmd.PreRunner) *cobra.Command {
	demoStopCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "stop [demo]",
			Short: "Stop a demo.",
			Args:  cobra.ExactArgs(1),
			RunE:  runDemoStopCommand,
		}, prerunner)

	return demoStopCommand.Command
}

func runDemoStopCommand(command *cobra.Command, args []string) error {
	ch := local.NewConfluentHomeManager()
	return run(ch, args[0], "stop.sh")
}

func fetchExamplesRepo(ch local.ConfluentHome) error {
	hasRepo, err := ch.HasFile("examples")
	if err != nil {
		return err
	}

	dir, err := ch.GetFile("examples")
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

	version, err := ch.GetConfluentVersion()
	if err != nil {
		return err
	}

	branch := plumbing.NewBranchReferenceName(fmt.Sprintf("%s-post", version))
	if err := tree.Checkout(&git.CheckoutOptions{Branch: branch}); err != nil {
		return err
	}

	return nil
}

func isSupported(demo string) bool {
	for _, supportedDemo := range supportedDemos {
		if demo == supportedDemo {
			return true
		}
	}
	return false
}

func run(ch local.ConfluentHome, demo string, script string) error {
	if !isSupported(demo) {
		return fmt.Errorf("demo not supported: %s", demo)
	}

	if err := fetchExamplesRepo(ch); err != nil {
		return err
	}

	scriptFile, err := ch.GetFile("examples", demo, script)
	if err != nil {
		return err
	}

	dir, err := ch.GetFile("examples", demo)
	if err != nil {
		return err
	}

	command := exec.Command(scriptFile)
	command.Dir = dir
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	return command.Run()
}
