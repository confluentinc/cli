# Confluent CLI

[![Build Status](https://semaphoreci.com/api/v1/projects/accef4bb-d1db-491f-b22e-0d438211c888/1992525/shields_badge.svg)](https://semaphoreci.com/confluent/cli)
![Release](release.svg)
[![codecov](https://codecov.io/gh/confluentinc/cli/branch/master/graph/badge.svg?token=67t1cdciLU)](https://codecov.io/gh/confluentinc/cli)

## Install

### Brew

Setup the Confluent Cloud brew tap:

    brew tap confluentinc/ccloud

For now, the CLI is private so you'll need to use your Github token:

    HOMEBREW_GITHUB_API_TOKEN=xxx brew install confluent-cli

### S3

We're publishing pre-built binaries to a private S3 bucket. Make sure you have your AWS
engineering creds setup in your `[default]` AWS profile locally, like normal.

To list all available packages for a version:

    VERSION=0.6.0
    aws s3 ls s3://cloud-confluent-bin/cli/${VERSION}/

To download the CLI for your OS and architecture:

    OS=Darwin
    ARCH=x86_64
    aws s3 cp s3://cloud-confluent-bin/cli/${VERSION}/confluent-cli_${VERSION}_${OS}_${ARCH}.tar.gz .

To install the CLI:

    tar -xzvf confluent-cli_${VERSION}_${OS}_${ARCH}.tar.gz
    sudo mv confluent-cli_${VERSION}_${OS}_${ARCH}/confluent* /usr/local/bin

Note: components must be installed in your `$PATH` for the CLI to pick them up.

## Developing

```
$ make compile-proto
$ go run main.go --help
```

The CLI automatically adds commands when their respective plugins are installed. Enabling the connect
commands by installing the plugins:

```
$ make install-plugins
```

Now you can run:

```
$ go run main.go connect list
```

# Packaging and Distribution

Either set the `GITHUB_TOKEN` environment variable or create `~/.config/goreleaser/github_token`
with this value. The token must have `repo` scope to deploy artifacts to Github.


