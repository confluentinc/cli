#!/bin/bash
# Test that Confluent Cloud login works in a production environment with SSO credentials.
# Usage: bash login-sso.sh <email>

echo -e "$1\n" | HOME=$(mktemp -d) $(find dist/ -name confluent) login headless-sso --url https://devel.cpdev.cloud
