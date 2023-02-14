# Confluent CLI

[![Release](release.svg)](https://github.com/confluentinc/cli/releases/latest)
[![Build Status](https://confluent-cli.semaphoreci.com/badges/cli/branches/main.svg?style=shields&key=d7163855-c2f5-40b9-a5d7-ff9e3e2214fe)](https://confluent-cli.semaphoreci.com/projects/cli)

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

(If the directory has insufficient permissions, you may need to prefix `sh` with `sudo`)

#### Install a Specific Version

To list all available versions:

    curl -sL https://cnfl.io/cli | sh -s -- -l

For example, to install version `v3.0.0`:

    curl -sL https://cnfl.io/cli | sh -s -- v3.0.0

### Download a Tarball from S3

To list all available versions:

    curl -s "https://s3-us-west-2.amazonaws.com/confluent.cloud?prefix=confluent-cli/archives/&delimiter=/" | tidy -xml --wrap 100 -i - 2>/dev/null

For example, to list all available packages for version v3.0.0:

    VERSION=v3.0.0 # or latest
    curl -s "https://s3-us-west-2.amazonaws.com/confluent.cloud?prefix=confluent-cli/archives/${VERSION#v}/&delimiter=/" | tidy -xml --wrap 100 -i - 2>/dev/null

For example, to download a tarball for Darwin/amd64:

    VERSION=v3.0.0 # or latest
    OS=darwin
    ARCH=amd64
    FILE=confluent_${VERSION#v}_${OS}_${ARCH}.tar.gz
    curl -s https://s3-us-west-2.amazonaws.com/confluent.cloud/confluent-cli/archives/${VERSION#v}/${FILE} -o ${FILE}

To install the CLI from a tarball:

    tar -xzvf ${FILE}
    mv confluent/confluent /usr/local/bin

### Building from Source

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

#### Troubleshooting

Please update your system to MacOS 11.0 or later if you are building on Darwin/arm64.
