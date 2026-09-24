#!/usr/bin/env bash
# Prints a `go test -run` pattern selecting shard <index> (1-based) of <count> CLITestSuite methods.
# Uses `\z` instead of `$` as the end anchor so the pattern survives Make variable expansion.
set -euo pipefail

index=$1
count=$2

cd "$(dirname "$0")"

methods=$(sed -nE 's/^func \([A-Za-z0-9_]+ \*CLITestSuite\) (Test[A-Za-z0-9_]+)\(.*/\1/p' ./*.go | sort)

# fail rather than silently drop a suite method declared in a shape the pattern above misses
declared=$(grep -hE '^func[[:space:]]*\([^)]*CLITestSuite\)[[:space:]]*Test' ./*.go | wc -l)
if [ "$(echo "$methods" | wc -l)" -ne "$declared" ]; then
	echo "shard-tests.sh: extracted $(echo "$methods" | wc -l) of $declared CLITestSuite test methods" >&2
	exit 1
fi

# the pattern only reaches tests nested under TestCLI, and `go list` also selects packages below
# test/ (test/live is excluded by its build tag)
if grep -rhE --include='*.go' --exclude-dir=live --exclude-dir=fixtures '^func Test' . | grep -qv '^func TestCLI('; then
	echo "shard-tests.sh: found a top-level test other than TestCLI, which no shard would run" >&2
	exit 1
fi

shard=$(echo "$methods" | awk -v i="$index" -v n="$count" '(NR - 1) % n == i - 1' | paste -sd '|' -)

echo "^TestCLI\\z/^(${shard})\\z"
