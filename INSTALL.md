# Installing the Confluent Cloud CLI

Download the latest archive distribution for Mac OS X, Windows, or Linux
from the Confluent Cloud CLI S3 repository and install it on your system.

The specific procedures vary by operating system, but the following example
illustrates downloading and installing the binaries on Mac OS X.

1. Download and extract the Mac OS X binaries:

        $ curl -L -o ccloud-cli.tar.gz https://s3-us-west-2.amazonaws.com/confluent.cloud/ccloud-cli/archives/0.26.0-alpha2/ccloud_SNAPSHOT-f044964_darwin_amd64.tar.gz
        $ mkdir ccloud-cli && tar -xvzf ccloud-cli.tar.gz -C ccloud-cli

2. Move the binaries (`ccloud`, `ccloud-*-plugin`) to `/usr/local/bin`, or another location in your `$PATH`:

        $ mv ccloud-cli/ccloud* /usr/local/bin

3. Confirm your ccloud CLI version.

        $ ccloud version
