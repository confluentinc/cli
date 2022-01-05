# Confluent CLI

[![Build Status](https://dev.azure.com/confluentinc/cli/_apis/build/status/confluentinc.cli?branchName=master)](https://dev.azure.com/confluentinc/cli/_build/latest?definitionId=1&branchName=master)
![Release](release.svg)

The Confluent CLI lets you manage your Confluent Cloud and Confluent Platform deployments, right from the terminal.

## Install

The CLI has pre-built binaries for macOS, Linux, and Windows.

You can download a tarball with the binaries. These are both on Github releases and in S3.

### One Liner

The simplest way to install cross platform is with this one-liner:

    curl -sL https://cnfl.io/cli | sh

It'll install in `./bin` by default.

#### Install Dir

You can also install to a specific directory. For example, install to `/usr/local/bin` by running:

    curl -sL https://cnfl.io/cli | sudo sh -s -- -b /usr/local/bin

#### Install Version

You can list all available versions:

    curl -sL https://cnfl.io/cli | sh -s -- -l

And install a particular version if you desire:

    curl -sL https://cnfl.io/cli | sudo sh -s -- v0.64.0

This downloads a binary tarball from S3 compiled for your distro and installs it.

### Binary Tarball from S3

You can download a binary tarball from S3.

To list all available versions:

    curl -s "https://s3-us-west-2.amazonaws.com/confluent.cloud?prefix=confluent-cli/archives/&delimiter=/" | tidy -xml --wrap 100 -i - 2>/dev/null

To list all available packages for a version:

    VERSION=v0.95.0 # or latest
    curl -s "https://s3-us-west-2.amazonaws.com/confluent.cloud?prefix=confluent-cli/archives/${VERSION#v}/&delimiter=/" | tidy -xml --wrap 100 -i - 2>/dev/null

To download a tarball for your OS and architecture:

    VERSION=v0.95.0 # or latest
    OS=darwin
    ARCH=amd64
    FILE=confluent_${VERSION}_${OS}_${ARCH}.tar.gz
    curl -s https://s3-us-west-2.amazonaws.com/confluent.cloud/confluent-cli/archives/${VERSION#v}/${FILE} -o ${FILE}

To install the CLI:

    tar -xzvf ${FILE}
    sudo mv confluent/confluent /usr/local/bin

### Building From Source

```
$ make deps
$ make build
$ dist/confluent_$(go env GOOS)_$(go env GOARCH)/confluent -h
```

Please update your system to MacOS 11.0 or later, if you are developing for Darwin/arm64.

If `make deps` fails with an "unknown revision" error, you probably need to put your username and a
[github personal access token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token)
in your ~/.netrc file as outlined [here](https://gist.github.com/technoweenie/1072829). The access token needs to be
[authorized](https://docs.github.com/en/github/authenticating-to-github/authorizing-a-personal-access-token-for-use-with-saml-single-sign-on) for SSO.

## Developing

### Go Version

This repo requires go1.17.5. We recommend you use [goenv](https://github.com/syndbg/goenv) to manage your Go versions.
There's a `.go-version` file in this repo with the exact version we use (and test against in CI).

We recommend cloning the `goenv` repo directly to ensure that you have access to the latest version of Go. If you've
already installed `goenv` with brew, uninstall it first:

    brew uninstall goenv

Now, clone the `goenv` repo:

    git clone https://github.com/syndbg/goenv.git ~/.goenv

Then, add the following to your shell profile:

    export GOENV_ROOT="$HOME/.goenv"
    export PATH="$GOENV_ROOT/bin:$PATH"
    eval $(goenv init -)
    export PATH="$PATH:$GOPATH/bin"

Finally, you can install the version of Go we use by running the following command inside the root directory:

    goenv install

### Mac Setup Notes

Our integration tests (`make test`) open a lot of files while they are running.
On MacOS, the default maximum number of open files is 256, which is too small
(you will see an error like `error retrieving command exit code` or
`too many open files`).  Thus, if you are setting up your development
environment on MacOS, run the following three commands *then restart*:

    echo 'kern.maxfiles=20480' | sudo tee -a /etc/sysctl.conf
    echo -e 'limit maxfiles 8192 20480\nlimit maxproc 1000 2000' | sudo tee -a /etc/launchd.conf
    echo 'ulimit -n 4096' | sudo tee -a /etc/profile

Remember to restart for these changes to take effect.

### File Layout

This repo follows the [golang standard project layout](https://github.com/golang-standards/project-layout).

Here's the basic file structure:

* `cmd/confluent/main.go` - entrypoint for the CLI binary
* `internal/cmd/command.go` - bootstraps the root `confluent` CLI command
* `internal/cmd/<command>/<command>.go` - defines each command we support

Things under `internal/cmd` are commands, things under `internal/pkg` are packages to be used by commands.

When you add a new command or resource, assuming it's already in the SDK, you generally just need to create
* `internal/cmd/<command>/<command>.go` (and test)

### Cross Compile for Other Platforms

Cross compilation from a Darwin/amd64 machine to Darwin/arm64, Linux/amd64 and Windows/amd64 platforms is supported. To build for Darwin/arm64, run the following:

    GOARCH=arm64 make cross-build

To cross compile for Linux (glibc or musl), install cross compiler `musl-cross` with homebrew:

    brew install FiloSottile/musl-cross/musl-cross
    GOOS=linux make cross-build

To cross compile for Windows/amd64, install `Mingw-w64` compilers with homebrew:

    brew install mingw-w64
    GOOS=windows make cross-build

Cross compilation from M1 Macbook (Darwin/arm64) to other platforms is also supported. For detailed documentation, refer to [How to Build CLI with Confluent-Kafka-go for All Platforms](https://confluentinc.atlassian.net/wiki/spaces/Foundations/pages/2610299218/How+to+Build+CLI+with+Confluent-Kafka-go+for+All+Platforms)

### URLs
Use the `login` command with the `--url` option to point to a different development environment

### Security
We use precommit hooks and Gitleaks to prevent potential secrets from being committed to this repo. 
Please enable precommit hooks:

* follow this [link](https://confluentinc.atlassian.net/wiki/spaces/trustsecurity/pages/2564260728/Pre-commit+Secret+Detection+Hooks#Pre-commitSecretDetectionHooks-InstallPre-commitcommand) 
  to install precommit tool locally.  
* go to the working directory of this repo
* Run `pre-commit install`


## Installer

This repo contains an [installer script](install.sh). It was
generated by [godownloader](https://github.com/goreleaser/godownloader) and
manually modified to download from S3 instead of GitHub. In turn, godownloader
relies on portable shell functions from [shlib](https://github.com/client9/shlib).

Although it's been manually modified, it's a fairly clean and simple bash script.
The major modifications include:
* reworked `github_release` into `s3_release` function
* updated `TARBALL_URL` and `CHECKSUM_URL` to point to S3 instead of GitHub API
* added a new `-l` flag to list versions from S3, since we can't link to our (private) GitHub repo
* extracted a `BINARY` variable instead of having binary names hardcoded in `execute`
* updated version/tag handling of the `v` prefix; it's expected in GitHub and inconsistently used in S3
* updated the usage message, logging, and file comments a bit

## Documentation

The [Confluent CLI Command Reference](https://docs.confluent.io/confluent-cli/current/command-reference/index.html)
is programmatically generated from the Cobra commands in this repo.

Just run:

    make docs

## Testing

The CLI is tested with a combination of unit tests and integration tests
(backed by mocks). These are both contained within this repo.

We also have end-to-end system tests for
* cloud-only functionality - [cc-system-tests](https://github.com/confluentinc/cc-system-tests/blob/master/test/cli_test.go)
* on-prem-only functionality - [muckrake](https://github.com/confluentinc/muckrake) (TODO: fix link to CLI tests)

To run all tests

    make test

UNIT_TEST_ARGS environment variable is used to manipulate unit test execution,
while INT_TEST_ARGS environment variable is for integration tests.

For example you can filter for a subset of unit test and a subset integration tests to be run

    make test UNIT_TEST_ARGS="-run TestApiTestSuite" INT_TEST_ARGS="-run TestCLI/Test_Confluent_Iam_RoleBinding_List"

More details on the use of these environment variables in the *Unit Test* and *Integration Test* sections.

### Unit Tests

Unit tests exist in `_test.go` files alongside the main source code files.

You can run the all unit tests with

    make unit-test

To run only a subset of unit tests, you must find the suite and test name and filter with

    # all tests within a suite
    make unit-test UNIT_TEST_ARGS="-run TestApiTestSuite"

    # a very specific subset of tests
    make unit-test UNIT_TEST_ARGS="-run TestApiTestSuite/TestCreateCloudAPIKey"

UNIT_TEST_ARGS is can also be used with `make test` target, if you want to filter unit tests but still run integration tests

    make test UNIT_TEST_ARGS="-run TestApiTestSuite/TestCreateCloudAPIKey"

### Integration Tests

The [./test](./test) directory contains the integration tests. These build a CLI
binary and invoke commands on it. These CLI integration tests roughly follow this
[pattern](http://lucapette.me/writing-integration-tests-for-a-go-cli-application):

1. table tests for quickly testing a variety of CLI commands
1. golden files are expected output fixtures for spec compliance testing
1. http test server for stubbing the Confluent Platform Control Plane API

Read the [CLITest](./test/cli_test.go) configuration to get a better idea
about how to write and configure your own integration tests.

You can update the golden files from the current output with

    make int-test INT_TEST_ARGS="-update"

You can skip rebuilding the CLI if it already exists in `dist` with

    make int-test INT_TEST_ARGS="-no-rebuild"

You can mix and match these flags. To update the golden files without rebuilding, and log verbosely

    make int-test INT_TEST_ARGS="-update -no-rebuild -v"

To run a single test case (or all test cases with a prefix)

    # all integration tests
    make int-test INT_TEST_ARGS="-run TestCLI"

    # all subtests of this `TestIAMRBACRoleBindingListOnPrem` integration tests
    make int-test INT_TEST_ARGS="-run TestCLI/TestIAMRBACRoleBindingListOnPrem"

    # a very specific subset of tests
    make int-test INT_TEST_ARGS="-run TestCLI/TestIAMRBACRoleBindingListOnPrem/iam_rbac_role-binding_list_--kafka-cluster-id_CID_--principal_User:frodo"

INT_TEST_ARGS is can also be used with `make test` target, if you want to filter or update integration tests but still run unit tests

    make test INT_TEST_ARGS="-run TestCLI/TestIAMRBACRoleBindingListOnPrem"


## Adding a New Command to the CLI

### Command Overview
Commands in the CLI follow the following syntax:

`confluent <resource> [subresource] <standard-verb> [args]`

We'll be implementing a `confluent config file show <num-times>` command that outputs the config file of the CLI a specified number of times. For example, `confluent config file 3` might output:
```
~/.confluent/config.json
~/.confluent/config.json
~/.confluent/config.json
```

### Creating the command file

Like all commands, this command will reside in `internal/cmd`. The directory name of the command should match the top-level resource, `config`, in this case. That directory already exists, but we'll need to create a file under the `config` directory, whose name matches the subresource of the command (in this case, `file.go`). We'll add the following code to it:

```go
package config

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
    pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type fileCommand struct {
	*pcmd.CLICommand
	prerunner pcmd.PreRunner
	analytics analytics.Client
}

func NewFileCommand(prerunner pcmd.PreRunner, analytics analytics.Client) *cobra.Command {
	cliCmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "file",
			Short: "Manage the config file.",
		}, prerunner)
	cmd := &fileCommand{
		CLICommand: cliCmd,
		prerunner:  prerunner,
		analytics:  analytics,
	}
	cmd.init()
	return cmd.Command
}

func (c *fileCommand) init() {
	showCmd := &cobra.Command{
		Use:   "show <num-times>",
		Short: "Show the config file a specified number of times.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.show),
	}
	c.AddCommand(showCmd)
}

func (c *fileCommand) show(cmd *cobra.Command, args []string) error {
	numTimes, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	filename := c.CLICommand.Config.Config.Filename
	if filename == "" {
		return errors.New(errors.EmptyConfigFileErrorMsg)
	}
	for i := 0; i < numTimes; i++ {
		utils.Println(cmd, filename)
	}
	return nil
}
```

#### `New[Command]` Function
Here, we create the actual Cobra top-level command `file`, specifying the syntax with `Use`, and a short description with `Short`. Then we initialize the command using `init`, a convention used in the CLI codebase. Since the CLI commands often require additional parameters, we use a wrapper around Cobra commands, in this case named `fileCommand`.

#### `init` Function
Here, we add the subcommands, in this case just `show`. We specify the usage messages, number of arguments our command needs, and the function that will be executed when our command is run. Note that all `RunE` function must be intialized using `cmd` package's `NewCLIRunE` function, which handles the common logic for all CLI commands.
#### Main (Work) Function
This function is named after the verb component of the command, `show`. It does the "heavy" lifting by parsing the `<num-times>` arg, retrieving the filename, and either printing its name to the console, or returning an error if there's no filename set.

#### Error Handling
See [error.md](errors.md) for details.

### Registering the Command
We must register our newly created command with the top-level `config` command located at `internal/cmd/config/command.go`. We add it to the `config` command with `c.AddCommand(NewFileCommand(c.prerunner, c.analytics))`.

With an entirely new command, we would also need to register it with the base top-level command (`confluent`) located at `internal/cmd/command.go`, using the same `AddCommand` syntax. Since the `config` is already registered, we can skip this step.

### Building
To build the CLI binary, we run `make build`. After this, we can run our command in the following way, and see that it (hopefully) works!

```
dist/confluent_<platform>_<arch>/confluent config file show 3
```

### Integration Testing
There's not much code here to unit test, so we'll skip right to integration testing. We'll create a file named `file_test.go` under the `test` directory, and add the following code to it:

```go
package test

func (s *CLITestSuite) TestFileCommands() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{name: "succeed if showing existing config file", args: "config file show 3", fixture: "file1.golden"},
	}
	resetConfiguration(s.T())

	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.workflow = true
		s.runConfluentTest(tt)
	}
}
```

We'll also need to add the new golden file, `file1.golden`, to `test/fixtures/output`. After running the command manually to ensure the output is correct, the content for the golden file can either be:

1. Copied directly from the shell
2. Generated automatically by running `make test INT_TEST_ARGS="-update"`, which runs all integration tests and updates all golden files to match their output. This is a risky command to run, as it essentially passes all integration tests, but is convenient to use if you can't get tests to pass from manual copying due to some hidden spaces. In addition to auto-filling the `file` golden file, this command will update the `help` command test outputs to reflect the added command.

To run this integration test, run `make test INT_TEST_ARGS="-run TestCLI/TestFileCommands"`.

### Opening a PR!

That's it! As you can see, the process of adding a new command is pretty straightforward. After you're able to successfully build the CLI with `make build`, and all unit and integration tests pass with `make test`, you can open a PR!
