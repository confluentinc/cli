#!/usr/bin/env bash
# Installs the Confluent CLI from the published YUM repo, following the steps in README.md.
# Runs inside a RHEL-compatible container as root; see `make install-smoke-rpm`.
# EXPECTED_VERSION (e.g. v4.78.0) is required; the installed version must match it.
set -euo pipefail

: "${EXPECTED_VERSION:?must be set, e.g. v4.78.0}"

rpm --import https://packages.confluent.io/confluent-cli/rpm/archive.key
yum install -y yum-utils
yum-config-manager --add-repo https://packages.confluent.io/confluent-cli/rpm/confluent-cli.repo
yum clean all && yum install -y confluent-cli

confluent version
help=$(confluent --help 2>&1)
grep -q "Manage your .*" <<< "${help}" || { echo "Unable to execute installed confluent CLI"; exit 1; }

installed=$(confluent version | sed -n 's/^Version:[[:space:]]*//p')
[ "${installed}" = "${EXPECTED_VERSION}" ] || { echo "Installed ${installed}, expected ${EXPECTED_VERSION}"; exit 1; }

echo "RPM install smoke test passed!"
