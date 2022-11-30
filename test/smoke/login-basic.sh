#!/bin/bash
# Test that Confluent Cloud login works in a production environment with basic (username and password) credentials.
# Usage: bash login-basic.sh <email> <password>

echo -e "$1\n$2\n" | HOME=$(mktemp -d) $(find dist/ -name confluent) login
