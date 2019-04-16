# Confluent Cloud CLI

[![Build Status](https://semaphoreci.com/api/v1/projects/accef4bb-d1db-491f-b22e-0d438211c888/1992525/shields_badge.svg)](https://semaphoreci.com/confluent/cli)
![Release](release.svg)
[![codecov](https://codecov.io/gh/confluentinc/cli/branch/master/graph/badge.svg?token=67t1cdciLU)](https://codecov.io/gh/confluentinc/cli)

This is the v2 Confluent *Cloud CLI*. It also serves as the backbone for the Confluent "*Converged CLI*" efforts.
In particular, the repository also contains all of the code for the on-prem "*Confluent CLI*", which is also built
as part of the repo's build process.

## Install

The CLI has pre-built binaries for mac, linux, and windows, on both i386 and x86_64 architectures.

You can download a tarball with the binaries. These are both on Github releases and in S3.

### Binary Tarball from S3

You can download a binary tarball from S3 .

To list all available versions:

    curl -s "https://s3-us-west-2.amazonaws.com/confluent.cloud?prefix=ccloud-cli/archives/&delimiter=/" | tidy -xml --wrap 100 -i -


To list all available packages for a version:

    VERSION=0.26.0 # or latest
    curl -s "https://s3-us-west-2.amazonaws.com/confluent.cloud?prefix=ccloud-cli/archives/${VERSION}/&delimiter=/" | tidy -xml --wrap 100 -i -

To download a tarball for your OS and architecture:

    VERSION=0.26.0 # or latest
    OS=darwin
    ARCH=amd64
    FILE=ccloud_v${VERSION}_${OS}_${ARCH}.tar.gz
    curl -s https://s3-us-west-2.amazonaws.com/confluent.cloud/ccloud-cli/archives/${VERSION}/${FILE} -o ${FILE}

To install the CLI:

    mkdir ccloud-cli && tar -xzvf ccloud_v${VERSION}_${OS}_${ARCH}.tar.gz -C ccloud-cli
    sudo mv ccloud-cli/ccloud* /usr/local/bin

To use the AWS S3 CLI instead of curl requires read access to Confluent Cloud AWS Prod account.
This is where the `confluent.cloud` S3 bucket is located.

### Building From Source

```
$ make deps
$ make build
$ dist/ccloud/$(go env GOOS)_$(go env GOARCH)/ccloud -h # for cloud CLI
$ dist/confluent/$(go env GOOS)_$(go env GOARCH)/confluent -h # for on-prem Confluent CLI
```

## Developing

This repo requires golang 1.11 and follows the basic
[golang standard project layout](https://github.com/golang-standards/project-layout).

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

## Testing

The CLI is tested with a combination of unit tests and integration tests
(backed by mocks). These are both contained within this repo.

We also have end-to-end system tests for
* ccloud-only functionality - [cc-system-tests](https://github.com/confluentinc/cc-system-tests/blob/master/test/cli_test.go)
* on-prem-only functionality - [muckrake](https://github.com/confluentinc/muckrake) (TODO: fix link to CLI tests)

Unit tests exist in `_test.go` files alongside the main source code files.

### Integration Tests

The [./test](./test) directory contains the integration tests. These build a CLI
binary and invoke commands on it. These CLI integration tests roughly follow this
[pattern](http://lucapette.me/writing-integration-tests-for-a-go-cli-application):

1. table tests for quickly testing a variety of CLI commands
1. golden files are expected output fixtures for spec compliance testing
1. http test server for stubbing the Confluent Platform Control Plane API

Read the [CLITest](./test/cli_test.go) configuration to get a better idea
about how to write and configure your own integration tests.

You can run just the integration tests with

    make test TEST_ARGS="./test/... -v"

You can update the golden files from the current output with

    make test TEST_ARGS="./test/... -update"

You can skip rebuilding the CLI if it already exists in `dist` with

    make test TEST_ARGS="./test/... -update -no-rebuild -v"
