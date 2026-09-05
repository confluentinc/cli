[![Try Confluent Cloud - The Data Streaming Platform](https://images.ctfassets.net/8vofjvai1hpv/10bgcSfn5MzmvS4nNqr94J/af43dd2336e3f9e0c0ca4feef4398f6f/confluent-banner-v2.svg)](https://confluent.cloud/signup?utm_source=github&utm_medium=banner&utm_campaign=oss-repos&utm_term=cli)

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

Or, optionally install the FIPS-140 compatible version of `confluent`:

    brew install confluentinc/tap/cli-fips

Then, follow the instructions below in the section titled "Build an OpenSSL FIPS Provider for FIPS-140 Mode".

#### APT (Ubuntu and Debian)

Install the latest version of `confluent` to `/usr/bin` (requires `glibc 2.28` or above):

    wget -qO - https://packages.confluent.io/confluent-cli/deb/archive.key | sudo apt-key add -
    sudo apt install software-properties-common
    sudo add-apt-repository "deb https://packages.confluent.io/confluent-cli/deb stable main"
    sudo apt update && sudo apt install confluent-cli

#### YUM (RHEL and CentOS)

Install the latest version of `confluent` to `/usr/bin` (requires `glibc 2.28` or above):

    sudo rpm --import https://packages.confluent.io/confluent-cli/rpm/archive.key
    sudo yum install yum-utils
    sudo yum-config-manager --add-repo https://packages.confluent.io/confluent-cli/rpm/confluent-cli.repo
    sudo yum clean all && sudo yum install confluent-cli

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

    make build
    dist/confluent_$(go env GOOS)_$(go env GOARCH)/confluent -h

#### Cross Compile for Other Platforms

From darwin/amd64 or darwin/arm64, you can build the CLI for any other supported platform.

To build for darwin/amd64 from darwin/arm64, run the following:

    GOARCH=amd64 make build

To build for darwin/arm64 from darwin/amd64, run the following:

    GOARCH=arm64 make build

To build for linux/amd64 (glibc or musl), run the following:

    brew install FiloSottile/musl-cross/musl-cross
    GOOS=linux GOARCH=amd64 make cross-build

To build for linux/arm64 (glibc or musl), run the following:

    brew install FiloSottile/musl-cross/musl-cross
    GOOS=linux GOARCH=arm64 make cross-build

To build for windows/amd64, run the following:

    brew install mingw-w64
    GOOS=windows GOARCH=amd64 make cross-build

#### Building for macOS in FIPS-140 mode

Linux is built in FIPS-140 mode by default. To build the CLI for macOS in FIPS-140 mode, set the `GOLANG_FIPS` environment variable to "1":

```bash
GOLANG_FIPS=1 make build
```

Then, follow the instructions in the next section to build an OpenSSL FIPS provider.

### Build an OpenSSL FIPS Provider for FIPS-140 Mode

```bash
wget "https://www.openssl.org/source/openssl-3.0.9.tar.gz"
tar -xvf openssl-3.0.9.tar.gz
cd openssl-3.0.9/
./Configure enable-fips
make install_fips DESTDIR=install
```

Copy the generated files into the Homebrew OpenSSL directory:

```bash
cp install/usr/local/lib/ossl-modules/fips.dylib /opt/homebrew/Cellar/openssl@3/<version>/lib/ossl-modules
cp install/usr/local/ssl/fipsmodule.cnf /opt/homebrew/etc/openssl@3/
```

Create a new OpenSSL configuration file for FIPS-140 mode:

```bash
cp /opt/homebrew/etc/openssl@3/openssl.cnf /opt/homebrew/etc/openssl@3/openssl-fips.cnf
```

Append the following to `openssl-fips.cnf`:

```
config_diagnostics = 1
openssl_conf = openssl_init

.include /opt/homebrew/etc/openssl@3/fipsmodule.cnf

[openssl_init]
providers = provider_sect
ssl_conf = ssl_module
alg_section = algorithm_sect

[provider_sect]
fips = fips_sect
default = default_sect

[default_sect]
activate = 1

[algorithm_sect]
default_properties = fips=yes

[ssl_module]
system_default = crypto_policy

[crypto_policy]
CipherString = @SECLEVEL=2:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384
Ciphersuites = TLS_AES_256_GCM_SHA384
TLS.MinProtocol = TLSv1.2
TLS.MaxProtocol = TLSv1.3
DTLS.MinProtocol = DTLSv1.2
DTLS.MaxProtocol = DTLSv1.2
SignatureAlgorithms = ECDSA+SHA256:ECDSA+SHA384:ECDSA+SHA512:rsa_pss_pss_sha256:rsa_pss_pss_sha384:rsa_pss_pss_sha512:rsa_pss_rsae_sha256:rsa_pss_rsae_sha384:rsa_pss_rsae_sha512:RSA+SHA256:RSA+SHA384:RSA+SHA512:ECDSA+SHA224:RSA+SHA224
```

Run the Confluent CLI in FIPS-140 mode:

```bash
env \
 DYLD_LIBRARY_PATH=/opt/homebrew/Cellar/openssl@3/<version>/lib \
 OPENSSL_CONF=/opt/homebrew/etc/openssl@3/openssl-fips.cnf \
 GOLANG_FIPS=1 \
 confluent version
```
