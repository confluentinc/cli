# Confluent CLI

[![Release](release.svg)](https://github.com/confluentinc/cli/releases/latest)
[![Build Status](https://dev.azure.com/confluentinc/cli/_apis/build/status/confluentinc.cli?branchName=main)](https://dev.azure.com/confluentinc/cli/_build/latest?definitionId=1&branchName=master)

The Confluent CLI lets you manage your Confluent Cloud and Confluent Platform deployments, right from the terminal.

## Documentation

The [Confluent CLI Documentation Website](https://docs.confluent.io/confluent-cli/current/overview.html) contains
detailed installation and setup information.

The [Confluent CLI Command Reference](https://docs.confluent.io/confluent-cli/current/command-reference/index.html)
contains information on command arguments and flags, and is programmatically generated from this repository.

## Contributing

All contributions are appreciated, no matter how small!
When opening a PR, please make sure to follow our [contribution guide](CONTRIBUTING.md).

## Installation

The Confluent CLI is available to install on macOS, Linux, and Windows.

### One-Liner

The simplest way to install the Confluent CLI is with this one-liner:

    curl -sL https://cnfl.io/cli | sh

(By default, the CLI will be installed in `./bin`)

#### Install to a Specific Directory

For example, to install to `/usr/local/bin`:

    curl -sL https://cnfl.io/cli | sh -s -- -b /usr/local/bin

#### Install a Specific Version

To list all available versions:

    curl -sL https://cnfl.io/cli | sh -s -- -l

For example, to install version `v2.3.1`:

    curl -sL https://cnfl.io/cli | sh -s -- v2.3.1

### Download a Tarball from S3

To list all available versions:

    curl -s "https://s3-us-west-2.amazonaws.com/confluent.cloud?prefix=confluent-cli/archives/&delimiter=/" | tidy -xml --wrap 100 -i - 2>/dev/null

For example, to list all available packages for version v2.3.1:

    VERSION=v2.3.1 # or latest
    curl -s "https://s3-us-west-2.amazonaws.com/confluent.cloud?prefix=confluent-cli/archives/${VERSION#v}/&delimiter=/" | tidy -xml --wrap 100 -i - 2>/dev/null

For example, to download a tarball for Darwin/amd64:

    VERSION=v2.3.1 # or latest
    OS=darwin
    ARCH=amd64
    FILE=confluent_${VERSION}_${OS}_${ARCH}.tar.gz
    curl -s https://s3-us-west-2.amazonaws.com/confluent.cloud/confluent-cli/archives/${VERSION#v}/${FILE} -o ${FILE}

To install the CLI from a tarball:

    tar -xzvf ${FILE}
    mv confluent/confluent /usr/local/bin

### Building from Source

    make deps
    make build
    dist/confluent_$(go env GOOS)_$(go env GOARCH)/confluent -h

#### Cross Compile for Other Platforms

Cross compilation from a Darwin/amd64 machine to Darwin/arm64, Linux/amd64 and Windows/amd64 platforms is supported.
To build for Darwin/arm64, run the following:

    GOARCH=arm64 make cross-build

To build for Linux (glibc or musl), install cross compiler `musl-cross` with homebrew:

    brew install FiloSottile/musl-cross/musl-cross
    GOOS=linux make cross-build

To build for Windows/amd64, install `mingw-w64` compilers with homebrew:

    brew install mingw-w64
    GOOS=windows make cross-build

Cross compilation from an M1 Macbook (Darwin/arm64) to other platforms is also supported.
For detailed documentation, refer to [How to Build the CLI with confluent-kafka-go for All Platforms](https://confluentinc.atlassian.net/wiki/spaces/Foundations/pages/2610299218/How+to+Build+CLI+with+Confluent-Kafka-go+for+All+Platforms)

#### Troubleshooting

Please update your system to MacOS 11.0 or later if you are building on Darwin/arm64.

If `make deps` fails with an "unknown revision" error, you probably need to put your username and a [GitHub personal access token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token)
in your ~/.netrc file as outlined [here](https://gist.github.com/technoweenie/1072829).
The access token needs to be [authorized for SSO](https://docs.github.com/en/github/authenticating-to-github/authorizing-a-personal-access-token-for-use-with-saml-single-sign-on).
