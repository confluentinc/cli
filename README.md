# Confluent Cloud CLI

[![Build Status](https://dev.azure.com/confluentinc/cli/_apis/build/status/confluentinc.cli?branchName=master)](https://dev.azure.com/confluentinc/cli/_build/latest?definitionId=1&branchName=master)
![Release](release.svg)
[![codecov](https://codecov.io/gh/confluentinc/cli/branch/master/graph/badge.svg?token=67t1cdciLU)](https://codecov.io/gh/confluentinc/cli)

This is the v2 Confluent *Cloud CLI*. It also serves as the backbone for the Confluent "*Converged CLI*" efforts.
In particular, the repository also contains all of the code for the on-prem "*Confluent CLI*", which is also built
as part of the repo's build process.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Install](#install)
  - [One Liner](#one-liner)
    - [Install Dir](#install-dir)
    - [Install Version](#install-version)
  - [Binary Tarball from S3](#binary-tarball-from-s3)
  - [Building From Source](#building-from-source)
- [Developing](#developing)
  - [Go Version](#go-version)
  - [Mac Setup Notes](#mac-setup-notes)
  - [File Layout](#file-layout)
  - [Build Other Platforms](#build-other-platforms)
  - [URLS](#urls)
  - [Releasing a New CLI Version](#releasing-a-new-cli-version)
- [Installers](#installers)
- [Documentation](#documentation)
  - [README](#readme)
  - [Reference Docs](#reference-docs)
- [Testing](#testing)
  - [Integration Tests](#integration-tests)
- [Adding a New Command to the CLI](#adding-a-new-command-to-the-cli)
  - [Command Overview](#command-overview)
  - [Creating the command file](#creating-the-command-file)
    - [`New[Command]` Function](#newcommand-function)
    - [`init` Function](#init-function)
    - [Main (Work) Function](#main-work-function)
  - [Registering the Command](#registering-the-command)
  - [Building](#building)
  - [Integration Testing](#integration-testing)
  - [Opening a PR!](#opening-a-pr)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Install

The CLI has pre-built binaries for macOS, Linux, and Windows, on both i386 and x86_64 architectures.

You can download a tarball with the binaries. These are both on Github releases and in S3.

### One Liner

The simplest way to install cross platform is with this one-liner:

    curl -sL https://cnfl.io/ccloud-cli | sh

Or for the on-prem binary (while these are separate binaries):

    curl -sL https://cnfl.io/cli | sh

It'll install in `./bin` by default.

#### Install Dir

You can also install to a specific directory. For example, install to `/usr/local/bin` by running:

    curl -sL https://cnfl.io/ccloud-cli | sudo sh -s -- -b /usr/local/bin

#### Install Version

You can list all available versions:

    curl -sL https://cnfl.io/ccloud-cli | sh -s -- -l

    curl -sL https://cnfl.io/cli | sh -s -- -l

And install a particular version if you desire:

    curl -sL https://cnfl.io/ccloud-cli | sudo sh -s -- v0.64.0

This downloads a binary tarball from S3 compiled for your distro and installs it.

### Binary Tarball from S3

You can download a binary tarball from S3.

To list all available versions:

    curl -s "https://s3-us-west-2.amazonaws.com/confluent.cloud?prefix=ccloud-cli/archives/&delimiter=/" | tidy -xml --wrap 100 -i - 2>/dev/null

To list all available packages for a version:

    VERSION=v0.95.0 # or latest
    curl -s "https://s3-us-west-2.amazonaws.com/confluent.cloud?prefix=ccloud-cli/archives/${VERSION#v}/&delimiter=/" | tidy -xml --wrap 100 -i - 2>/dev/null

To download a tarball for your OS and architecture:

    VERSION=v0.95.0 # or latest
    OS=darwin
    ARCH=amd64
    FILE=ccloud_${VERSION}_${OS}_${ARCH}.tar.gz
    curl -s https://s3-us-west-2.amazonaws.com/confluent.cloud/ccloud-cli/archives/${VERSION#v}/${FILE} -o ${FILE}

To install the CLI:

    tar -xzvf ${FILE}
    sudo mv ccloud/ccloud /usr/local/bin

### Building From Source

```
$ make deps
$ make build
$ dist/ccloud/ccloud_$(go env GOOS)_$(go env GOARCH)/ccloud -h # for cloud CLI
$ dist/confluent/confluent_$(go env GOOS)_$(go env GOARCH)/confluent -h # for on-prem Confluent CLI
```

If `make deps` fails with an "unknown revision" error, you probably need to put your username and a
[github personal access token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token)
in your ~/.netrc file as outlined [here](https://gist.github.com/technoweenie/1072829). The access token needs to be
[authorized](https://docs.github.com/en/github/authenticating-to-github/authorizing-a-personal-access-token-for-use-with-saml-single-sign-on) for SSO.

## Developing

This repo requires golang 1.15.5. We recommend you use `goenv` to manage your go versions.
There's a `.go-version` file in this repo with the exact version we use (and test against in CI).

### Go Version

Fortunately `goenv` supports 1.15.5 already. If your `goenv` does not list this as an option,
you may have to build `goenv`'s `master` branch from source, which you can do with the
following instructions:

If you already have it via brew, uninstall it first:

    brew uninstall goenv

Now clone the repo and update your shell profile:

    git clone https://github.com/syndbg/goenv.git $GOPATH/src/github.com/syndbg/goenv
    echo 'export GOENV_ROOT="$GOPATH/src/github.com/syndbg/goenv"' >> ~/.bash_profile
    echo 'export PATH="$GOENV_ROOT/bin:$PATH"' >> ~/.bash_profile
    echo 'eval "$(goenv init -)"' >> ~/.bash_profile

Install the required version of `goreleaser`

    go get github.com/goreleaser/goreleaser@v0.116.0


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
* `internal/pkg/sdk/<resource>/<resource>.go` - a thin wrapper around `ccloud-sdk-go` to add logging and typed errors
   TODO: if we add logging and typed errors to the SDK, we might be able to drop the pkg/sdk stuff entirely.

Things under `internal/cmd` are commands, things under `internal/pkg` are packages to be used by commands.

When you add a new command or resource, assuming its already in the SDK, you generally just need to create
* `internal/cmd/<command>/<command>.go` (and test)
* `internal/pkg/sdk/<resource>/<resource>.go` (and test)

### Build Other Platforms

If you have a need to build a binary for a platform that is not the current one, use the following to target a different `.goreleaser-*` file matching the destined platform.

    make build-go GORELEASER_SUFFIX=-linux.yml   # build linux
    make build-go GORELEASER_SUFFIX=-mac.yml     # build mac
    make build-go GORELEASER_SUFFIX=-windows.yml # build windows

### URLS
Use the `login` command with the `--url` option to point to a different development environment

### Releasing a New CLI Version

0. Note: Due to the requirement of signing the macOS binaries, CLI releases
can currently only be run from a macOS host.  Binaries for all OS will be
generated and published from the macOS host.

1. To release a new CLI version, first ensure you have an Apple ID with
permissions for the Confluent organization.  Chris Toney / IT can set you up
with the appropriate permissions; cc @David Hyde, @srini or CLI or Release team
manager.  You will also need IT to give you the Developer ID Application cert
and private key (shared via a LastPass secret).  Chris Toney set that up.

2. Make sure you have `gon` installed and in your PATH.  Get it from
https://github.com/mitchellh/gon/releases

3. Edit your bash/zsh profile to export your Apple ID and password via the
`AC_USERNAME` and `AC_PASSWORD` variables.  It is recommended to not write
your Apple ID password in plaintext in this file; e.g. you can write
`export AC_PASSWORD=$(echo <your base64-encoded password> | base64 --decode)`.

4. Just in case, e.g. if you did Step 2 for the first time, run
`source ~/.bash_profile` or the equivalent for your shell/OS.

5. Next run `caasenv prod` (needed to upload files to the CLI S3 buckets).

6. Next make sure your git status is clean: `git status`.  You should also be
on `master` under normal circumstances.

7. With these steps completed you should be able to run `make release`
successfully.  If you see any errors during the release process, contact
a CLI team member.  It is possible to use the AWS CLI to remove any mistakenly-
uploaded binaries/archives (just remove the corresponding version folders).
You could also remove an accidentally-published tag on GitHub, but this isn't
essential.

Note: you can verify whether a macOS binary is signed and notarized correctly
by running `spctl -a -vvv -t install <binary name>`.  If all is good, you
should see output like
```
dist/ccloud/ccloud_darwin_amd64/ccloud: accepted
source=Notarized Developer ID
origin=Developer ID Application: Confluent, Inc. (RTSX8FNWR2)
```

## Installers

This repo contains installers for [install-ccloud.sh](./install-ccloud.sh) and
[install-confluent.sh](./install-confluent.sh). These were based on installers
generated by [godownloader](https://github.com/goreleaser/godownloader) and
manually modified to download from S3 instead of GitHub. In turn, godownloader
relies on portable shell functions from [shlib](https://github.com/client9/shlib).

Although they've been manually modified, they're fairly clean and simple bash scripts.
The major modifications include
* reworked `github_release` into `s3_release` function
* updated `TARBALL_URL` and `CHECKSUM_URL` to point to S3 instead of GitHub API
* added a new `-l` flag to list versions from S3, since we can't link to our (private) github repo
* extracted a `BINARY` variable instead of having binary names hardcoded in `execute`
* updated version/tag handling of the `v` prefix; its expected in GitHub and inconsistently used in S3
* updated the usage message, logging, and file comments a bit

## Documentation

### README

To re-generate the Table of Contents in this README, you'll need to install nodejs.
You can install a specific version [directly](https://nodejs.org/en/) or use
[nvm](https://github.com/nvm-sh/nvm) for multi-node version support (like goenv).

Then run this to update the Table of Contents:

    $ make doctoc

### Reference Docs

The CLI command [reference docs](https://docs.confluent.io/current/cloud/cli/command-reference/index.html)
are programmatically generated from the Cobra commands in this repo.

Just run:

    $ make docs

Cheat sheet:
```
	cli := &cobra.Command{
		Use:               cliName,
		Short:             "This is a short description",
		Long:              "This is a longer synopsis",
	}
```

## Testing

The CLI is tested with a combination of unit tests and integration tests
(backed by mocks). These are both contained within this repo.

We also have end-to-end system tests for
* ccloud-only functionality - [cc-system-tests](https://github.com/confluentinc/cc-system-tests/blob/master/test/cli_test.go)
* on-prem-only functionality - [muckrake](https://github.com/confluentinc/muckrake) (TODO: fix link to CLI tests)

To run all tests

    make test

UNIT_TEST_ARGS environment variable is used to manipulate unit test execution,
while INT_TEST_ARGS environment variable is for integration tests.

For example you can filter for a subset of unit test and a subset integration tests to be run

    make test UNIT_TEST_ARGS="-run TestApiTestSuite" INT_TEST_ARGS="-run TestCLI/Test_Confluent_Iam_Rolebinding_List"

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

    # all subtests of this `Test_Confluent_Iam_Rolebinding_List` integration tests
    make int-test INT_TEST_ARGS="-run TestCLI/Test_Confluent_Iam_Rolebinding_List"

    # a very specific subset of tests
    make int-test INT_TEST_ARGS="-run TestCLI/Test_Confluent_Iam_Rolebinding_List/iam_rolebinding_list_--kafka-cluster-id_CID_--principal_User:frodo"

INT_TEST_ARGS is can also be used with `make test` target, if you want to filter or update integration tests but still run unit tests

    make test INT_TEST_ARGS="-run TestCLI/Test_Confluent_Iam_Rolebinding_List"


## Adding a New Command to the CLI

### Command Overview
Commands in the CLI follow the following syntax:

`ccloud/confluent <resource> [subresource] <standard-verb> [args]`

We'll be implementing a `ccloud config file show <num-times>` command that outputs the config file of the CLI a specified number of times. For example, `ccloud config file 3` might output:
```
~/.ccloud/config.json
~/.ccloud/config.json
~/.ccloud/config.json
```

### Creating the command file

Like all commands, this command will reside in `internal/cmd`. The directory name of the command should match the top-level resource, `config`, in this case. That directory already exists, but we'll need to create a file under the `config` directory, whose name matches the subresource of the command (in this case, `file.go`). We'll add the following code to it:

```go
package config

import (
	"strconv"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

type fileCommand struct {
	*cobra.Command
	config *v3.Config
}

func NewFileCommand(config *v3.Config) *cobra.Command {
	cmd := &fileCommand{
		Command: &cobra.Command{
			Use:   "file",
			Short: "Manage the config file.",
		},
		config: config,
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
	filename := c.config.Filename
	if filename == "" {
		return errors.New(errors.NoConfigFileErrorMsg)
	}
	for i := 0; i < numTimes; i++ {
		pcmd.Println(cmd, filename)
	}
	return nil
}
```

#### `New[Command]` Function
Here, we create the actual Cobra top-level command `file`, specifying the syntax with `Use`, and a short description with `Short`. Then we initialize the command using `init`, a convention used in the CLI codebase. Since the CLI commands often require additional parameters, we use a wrapper around Cobra commands, in this case named `fileCommand`.

For our command, the constructor needs to take a `Config` struct as a parameter. `Config` describes the configuration of the CLI, and is parsed from a file located by default at `~/.ccloud/config.json` for `ccloud` commands, and at `~/.confluent/config.json` for `confluent` commands.



#### `init` Function
Here, we add the subcommands, in this case just `show`. We specify the usage messages, number of arguments our command needs, and the function that will be executed when our command is run. Not that all `RunE` function must be intialized using `cmd` package's `NewCLIRunE` function, which handles the common logic for all CLI commands.
#### Main (Work) Function
This function is named after the verb component of the command, `show`. It does the "heavy" lifting by parsing the `<num-times>` arg, retrieving the filename, and either printing its name to the console, or returning an error if there's no filename set.

#### Error Handling
See [error.md](errors.md) for details.

### Registering the Command
We must register our newly created command with the top-level `config` command located at `internal/cmd/config/command.go`. We add it to the `config` command with `c.AddCommand(NewFileCommand(c.config))`.

With an entirely new command, we would also need to register it with the base top-level command (`ccloud` and/or `confluent`) located at `internal/cmd/command.go`, using the same `AddCommand` syntax. Since the `config` is already registered, we can skip this step.

### Building
To build both binaries, we run `make build`. After this, we can run our command either of the following ways, and see that they (hopefully) work!

```
dist/ccloud/ccloud_<platform>/ccloud config file show 3
dist/confluent/confluent_<platform>/confluent config file show 3
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
	resetConfiguration(s.T(), "ccloud")
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.workflow = true
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runCcloudTest(tt, serve(s.T(), kafkaAPIURL).URL)
	}
}
```

We'll also need to add the new golden file, `file1.golden`, to `test/fixtures/output`. After running the command manually to ensure the output is correct, the content for the golden file can either be:

1. Copied directly from the shell
2. Generated automatically by running `make test INT_TEST_ARGS="-update"`, which runs all integration tests and updates all golden files to match their output. This is a risky command to run, as it essentially passes all integration tests, but is convenient to use if you can't get tests to pass from manual copying due to some hidden spaces. In addition to auto-filling the `file` golden file, this command will update the `help` command test outputs to reflect the added command.

To run this integration test, run `make test INT_TEST_ARGS="-run TestCLI/TestFileCommands"`.

### Opening a PR!

That's it! As you can see, the process of adding a new command is pretty straightforward. After you're able to successfully build the CLI with `make build`, and all unit and integration tests pass with `make test`, you can open a PR!
