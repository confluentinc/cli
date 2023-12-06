# Confluent CLI

[![Release](https://img.shields.io/github/v/release/confluentinc/cli)](https://github.com/confluentinc/cli/releases/latest)
[![Build Status](https://semaphore.ci.confluent.io/badges/cli/branches/main.svg?style=shields&key=36d1298e-932a-4d04-8cd0-2483a2a6ab85)](https://semaphore.ci.confluent.io/projects/cli)

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

#### Homebrew

Install the latest version of `confluent` to `/usr/local/bin`:

    brew install confluentinc/tap/cli

#### Scripted installation

Install the latest version of `confluent` to `/usr/local/bin`:

    curl -sL https://raw.githubusercontent.com/confluentinc/cli/main/install.sh | sh -s -- -b /usr/local/bin

Print a complete list of versions available for download:

    curl -sL https://raw.githubusercontent.com/confluentinc/cli/main/install.sh | sh -s -- -l

Install `confluent` v3.6.0 to `/usr/local/bin`:

    curl -sL https://raw.githubusercontent.com/confluentinc/cli/main/install.sh | sh -s -- -b /usr/local/bin 3.6.0

#### Windows

1. Download the latest Windows ZIP file from https://github.com/confluentinc/cli/releases/latest
2. Unzip `confluent_X.X.X_windows_amd64.zip`
3. Run `confluent.exe`

#### Docker

Pull the latest version:

    docker pull confluentinc/confluent-cli:latest

Pull `confluent` v3.6.0:

    docker pull confluentinc/confluent-cli:3.6.0

### Building from Source

    make
    dist/confluent_$(go env GOOS)_$(go env GOARCH)/confluent -h

#### Cross Compile for Other Platforms

Cross compilation from a Darwin/amd64 machine to Darwin/arm64, Linux/amd64 and Windows/amd64 platforms is supported.
To build for Darwin/arm64, run the following:

    GOARCH=arm64 make cross-build

To build for Linux (glibc or musl), install cross compiler `musl-cross` with Homebrew:

    brew install FiloSottile/musl-cross/musl-cross
    GOOS=linux make cross-build

To build for Windows/amd64, install `mingw-w64` compilers with Homebrew:

    brew install mingw-w64
    GOOS=windows make cross-build

Cross compilation from an M1 Macbook (Darwin/arm64) to other platforms is also supported.

#### Troubleshooting

Please update your system to MacOS 11.0 or later if you are building on Darwin/arm64.
