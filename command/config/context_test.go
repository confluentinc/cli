package config

import (
	"os"
	"testing"

	terminal "github.com/confluentinc/cli/command"
	"github.com/confluentinc/cli/shared"
	"github.com/stretchr/testify/require"
)

var (
	filename = "/tmp/cli-test-config"
)

func TestContext(t *testing.T) {
	req := require.New(t)
	os.Remove(filename)

	output, err := run("context", "current")
	req.NoError(err)
	req.Regexp(output, "^\n$")

	output, err = run("context", "list")
	req.NoError(err)
	req.Equal("  CURRENT | NAME | PLATFORM | CREDENTIAL  \n+---------+------+----------+------------+\n", output)

	output, err = run("context", "set", "my-context", "--kafka-cluster", "bob")
	req.NoError(err)
	req.Empty(output)

	output, err = run("context", "list")
	req.NoError(err)
	req.Equal("  CURRENT |    NAME    | PLATFORM | CREDENTIAL  \n+---------+------------+----------+------------+\n          | my-context |          |             \n", output)

	output, err = run("context", "get", "my-context")
	req.NoError(err)
	req.Contains(output, "credentials: \"\"\nkafka_cluster: bob\nplatform: \"\"\n\n")

	output, err = run("context", "get", "other-context")
	req.NoError(err)
	req.Contains(output, "credentials: \"\"\nkafka_cluster: \"\"\nplatform: \"\"\n\n")

	output, err = run("context", "list")
	req.NoError(err)
	req.NotContains(output,"other-context")

	output, err = run("context", "use", "my-context")
	req.NoError(err)
	req.Equal("", output)

	output, err = run("context", "current")
	req.NoError(err)
	req.Regexp(output, "^my-context\n$")

	os.Remove(filename)
}

func run(args ...string) (string, error) {
	config := shared.NewConfig()
	config.Filename = filename
	config.Load()
	root := New(config)

	return terminal.ExecuteCommand(root, args...)
}
