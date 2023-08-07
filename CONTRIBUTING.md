## Contributing

Thanks for your interest in contributing to the Confluent CLI!

### Development Environment

Start by following these steps to set up your computer for CLI development:

#### Go Version

We recommend you use [goenv](https://github.com/syndbg/goenv) to manage your Go versions.
There's a `.go-version` file in this repo with the exact version we use (and test against in CI).

We recommend cloning the `goenv` repo directly to ensure that you have access to the latest version of Go. If you've
already installed `goenv` with brew, uninstall it first:

    brew uninstall goenv

Now, clone the `goenv` repo:

    git clone https://github.com/syndbg/goenv.git ~/.goenv

Then, add the following to your shell profile:

    export GOENV_ROOT="$HOME/.goenv"
    export PATH="$GOENV_ROOT/bin:$PATH"
    eval "$(goenv init -)"
    export PATH="$PATH:$GOPATH/bin"

Finally, you can install the appropriate version of Go by running the following command inside the root directory of the repository:

    goenv install

#### Developing on MacOS

Our integration tests read a lot of files while they are running. On MacOS, the default maximum number of open files is
256, which is too small (you will see an error like `error retrieving command exit code` or `too many open files`).
Please run the following three commands *and then restart* for these changes to take effect:

    echo 'kern.maxfiles=20480' | sudo tee -a /etc/sysctl.conf
    echo -e 'limit maxfiles 8192 20480\nlimit maxproc 1000 2000' | sudo tee -a /etc/launchd.conf
    echo 'ulimit -n 4096' | sudo tee -a /etc/profile

#### Security

We use `pre-commit` hooks and `gitleaks` to prevent secrets from being committed to this repo. Please install `pre-commit` hooks (Note that the second command should be run inside the root directory of the repository):

    brew install pre-commit
    pre-commit install

### File Layout

This repo mostly follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout).
Here's the basic file structure:

    cli/
    ├─ cmd/
    │  ├─ confluent/
    │  │  ├─ main.go                            (entry point for the CLI binary)
    ├─ dist/
    │  ├─ confluent_<os>_<arch>/
    │  │  ├─ confluent                          (the CLI binary)
    ├─ internal/
    │  ├─ cmd/                                  (CLI commands)
    │  │  ├─ <command>/
    │  │  │  ├─ command.go                      (a top-level CLI command)
    │  │  │  ├─ command_<subcommand>.go         (a subcommand of a top-level CLI command)
    │  │  │  ├─ command_<subcommand>_onprem.go  (the on-prem version of the above command, if applicable)
    │  │  ├─ command.go                         (the root CLI command)
    │  ├─ pkg/
    ├─ test/                                    (integration tests)
    │  ├─ fixtures/
    │  │  ├─ output/
    │  │  │  ├─ <command>/                      (the golden files for a top-level CLI command)
    │  │  ├─ cli_test.go                        (entry point for all integration tests)
    │  │  ├─ <command>_test.go                  (the integration tests for a top-level CLI command)

### Testing

The CLI is tested with a combination of unit tests and integration tests. To run all tests:

    make test

#### Unit Tests

Unit tests exist in files ending with `_test.go`, and are located alongside the main source code files.
Unit tests should test small, isolated functions, and should not be unnecessarily complex (i.e. mocking backend calls or CLI commands).

To run the all unit tests:

    make unit-test

To run a subset of unit tests, you must specify the suite and optionally the name of a specific tests:

    # Run a suite of unit tests
    make unit-test UNIT_TEST_ARGS="-run TestApiTestSuite"

    # Run a specific unit test within a suite
    make unit-test UNIT_TEST_ARGS="-run TestApiTestSuite/TestCreateCloudAPIKey"

#### Integration Tests

The [test/](./test) directory contains our integration tests. These tests build the CLI binary and invoke commands on it.
These CLI integration tests roughly follow this [pattern](http://lucapette.me/writing-integration-tests-for-a-go-cli-application):

1. Run a test HTTP server to mock Confluent Cloud or the Confluent Platform Control Plane API.
2. Run a logical sequence of CLI commands.
3. Ensure that the output of these commands matches the corresponding golden files.

To update the golden files from the current output:

    make integration-test INTEGRATION_TEST_ARGS="-update"

To skip rebuilding the CLI, if it already exists in `dist/`:

    make integration-test INTEGRATION_TEST_ARGS="-no-rebuild"

To run a subset of integration tests, you must specify the suite and optionally the name of a specific test:

    # Run a suite of integration tests
    make integration-test INTEGRATION_TEST_ARGS="-run TestCLI/TestKafka"

    # Run a specific integration test within a suite
    make integration-test INTEGRATION_TEST_ARGS="-run TestCLI/TestKafka/kafka_cluster_--help"

### Example: Adding a New Command to the CLI

As a basic demonstration, we'll be implementing a command which prints the name of the CLI config file a specified number of times:

    $ confluent config describe 3
    ~/.confluent/config.json
    ~/.confluent/config.json
    ~/.confluent/config.json

#### Creating the Command

Like all other commands, this command will reside in `internal/cmd`. First, we must create a directory for this command:

    mkdir internal/cmd/config

Next, we create two files, one for the top-level command `config`, and another for `describe`.

`internal/cmd/config/command.go`:

```go
package config

import (
    "github.com/spf13/cobra"

    pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
    *pcmd.CLICommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
    cmd := &cobra.Command{Use: "config"}

    c := &command{pcmd.NewAnonymousCLICommand(cmd, prerunner)}

    cmd.AddCommand(c.newDescribeCommand())

    return cmd
}
```

`internal/cmd/config/command_describe.go`:

```go
package config

import (
    "strconv"

    "github.com/spf13/cobra"

    "github.com/confluentinc/cli/internal/pkg/errors"
    "github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDescribeCommand() *cobra.Command {
    return &cobra.Command{
        Use:  "describe",
        Args: cobra.ExactArgs(1),
        RunE: c.describe,
    }
}

func (c *command) describe(_ *cobra.Command, args []string) error {
    filename := c.CLICommand.Config.Config.Filename
    if filename == "" {
        return errors.New("config file not found")
    }
	
    n, err := strconv.Atoi(args[0])
    if err != nil {
        return err
    }

    for i := 0; i < n; i++ {
        output.Println(filename)
    }
    return nil
}

```

#### Registering the Command

Finally, we must add the newly created `config` command as a child of the root command.
Add the following line to `internal/cmd/command.go`, and make sure to import its package:

    cmd.AddCommand(config.New(prerunner))

#### Running the Command

To build the CLI binary, run `make build`. After this, we can run our command in the following way, and see that it (hopefully) works!

    make build
    dist/confluent_<os>_<arch>/confluent config file describe 3

#### Integration Testing

There's not much code here to unit test, so we'll skip right to integration testing. Create the following file:

`test/config/config_test.go`:

```go
package test

func (s *CLITestSuite) TestConfigDescribe() {
	tests := []CLITest{
		{args: "config describe 3", fixture: "config/1.golden"},
	}

	for _, test := range tests {
		s.runConfluentTest(test)
	}
}
```

We'll also need to add the new golden file, `test/fixtures/output/config/1.golden`.
After running the command manually to ensure the output is correct, the content for the golden file can either be:

1. Copied directly from the terminal.
2. Updated automatically with `make integration-test INTEGRATION_TEST_ARGS="-run TestCLI/TestConfigDescribe -update"` (slow).

Now, run `make integration-test INTEGRATION_TEST_ARGS="-run TestCLI/TestConfigDescribe"` and verify that it works!

#### Add Autocompletion

Add support for autocompletion using `ValidArgsFunction` if applicable (for example, if the command takes resource IDs or resource names as arguments):

```go
func (c *command) newDescribeCommand() *cobra.Command {
    return &cobra.Command{
        Use:               "describe",
        Args:              cobra.ExactArgs(1),
        ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
        RunE:              c.describe,
    }
}
```

See the [Autocompletion](internal/pkg/cmd/AUTOCOMPLETION.md) resource for implementation details.

### Opening a PR

That's it! As you can see, the process of adding a new CLI command is pretty straightforward. You can open a PR if:
* You're able to build the CLI with `make build`.
* All unit and integration test pass with `make test`.
* Running `make lint` produces no linter errors.

Note: If there is a JIRA ticket associated with your PR, please format the PR description as "[CLI-1234] Description of PR".

### Detailed Implementation Guides

Please familiarize yourself with the following resources before writing your first CLI command:

* [Cloud and On-Prem Annotations](internal/pkg/cmd/ANNOTATIONS.md)
* [CLI Error Handling](internal/pkg/errors/README.md)
* [Autocompletion](internal/pkg/cmd/AUTOCOMPLETION.md)
* TODO: REST Proxy Commands
