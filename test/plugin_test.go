package test

import (
	"github.com/stretchr/testify/require"
	"os"
)

func (s *CLITestSuite) TestPlugin() {
	path := os.Getenv("PATH")
	err := os.Setenv("PATH", "test/fixtures/input/plugin")
	require.NoError(s.T(), err)
	defer func() {
		err := os.Setenv("PATH", path)
		require.NoError(s.T(), err)
	}()
	//TODO: plugin that prints to stderr
	//TODO: handle plugin with a dash in name?
	tests := []CLITest{
		{args: "plugin1 arg1", fixture: "plugin/plugin1.golden"},
		{args: "print args arg1 arg2 --meaningless-flag=true arg3", fixture: "plugin/print-args.golden"},
		{args: "version", fixture: "plugin/exact-name-overlap.golden", regex: true},
		{args: "kafka something kafkaesque", fixture: "plugin/partial-name-overlap.golden"},
		{args: "foo bar baz boo far foo bar baz --flag=true", fixture: "plugin/long-plugin-name.golden"},
		//{args: "can print to stderr --meaningless-flag=false and stdout", fixture: "plugin/print-stderr.golden"},
		{args: "plugin list", fixture: "plugin/list.golden"},
	}

	resetConfiguration(s.T())

	for _, tt := range tests {
		tt.workflow = true
		s.runIntegrationTest(tt)
	}
}
