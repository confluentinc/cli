# Confluent CLI

[![Release](release.svg)](https://github.com/confluentinc/cli/releases/latest)
[![Build Status](https://confluent-cli.semaphoreci.com/badges/cli/branches/main.svg?style=shields&key=d7163855-c2f5-40b9-a5d7-ff9e3e2214fe)](https://confluent-cli.semaphoreci.com/projects/cli)

The Confluent CLI lets you manage your Confluent Cloud and Confluent Platform deployments, right from the terminal.

## Documentation

The [Confluent CLI Overview](https://docs.confluent.io/confluent-cli/current/overview.html) shows how to get started with the Confluent CLI.

The [Confluent CLI Command Reference](https://docs.confluent.io/confluent-cli/current/command-reference/index.html)
contains information on command arguments and flags, and is programmatically generated from this repository.

## Contributing

All contributions are appreciated, no matter how small!
When opening a PR, please make sure to follow our [contribution guide](CONTRIBUTING.md).

## Installation

The Confluent CLI is available to install for macOS, Linux, and Windows.

#### macOS

1. Download the latest macOS tar.gz file for your architecture type from https://github.com/confluentinc/cli/releases/latest
2. Unzip the file: `tar -xvf confluent_X.X.X_darwin_XXXXX.tar.gz`
3. Move `confluent` to a folder in your `$PATH`, such as `/usr/local/bin`

#### Linux

1. Download the latest Linux tar.gz file for your architecture type from https://github.com/confluentinc/cli/releases/latest
2. Unzip the file: `tar -xvf confluent_X.X.X_linux_XXXXX.tar.gz`
3. Move `confluent` to a folder in your `$PATH`

#### Windows

1. Download the latest Windows ZIP file from https://github.com/confluentinc/cli/releases/latest
2. Unzip `confluent_X.X.X_windows_amd64.zip`
3. Run the unzipped .exe file

#### Install a Specific Version

See the [releases page](https://github.com/confluentinc/cli/releases) for a complete list of versions available for download.

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

#### Troubleshooting

Please update your system to MacOS 11.0 or later if you are building on Darwin/arm64.
