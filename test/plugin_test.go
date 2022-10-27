package test

import (
	"os"
	"runtime"
	"strings"

	"github.com/stretchr/testify/require"
)

func (s *CLITestSuite) TestPlugin() {
	path := os.Getenv("PATH")
	newPath := "bin:test/fixtures/input/plugin:test/fixtures/input/plugin/test"
	if runtime.GOOS == "windows" {
		newPath = strings.ReplaceAll(newPath, ":", ";")
	}
	err := os.Setenv("PATH", newPath)
	require.NoError(s.T(), err)
	defer func() {
		err := os.Setenv("PATH", path)
		require.NoError(s.T(), err)
	}()

	tests := []CLITest{
		{args: "plugin1 arg1", fixture: "plugin/plugin1.golden", arePluginsEnabled: true},
		{args: "print args arg1 arg2 --meaningless-flag=true arg3", fixture: "plugin/print-args.golden", arePluginsEnabled: true},
		{args: "version", fixture: "plugin/exact-name-overlap.golden", regex: true, arePluginsEnabled: true},
		{args: "kafka something kafkaesque", fixture: "plugin/partial-name-overlap.golden", arePluginsEnabled: true},
		{args: "foo bar baz boo far foo bar baz --flag=true", fixture: "plugin/long-plugin-name.golden", arePluginsEnabled: true},
		{args: "can print to stderr --meaningless-flag=false and stdout", fixture: "plugin/print-stderr.golden", arePluginsEnabled: true},
		{args: "dash_test", fixture: "plugin/dash-test1.golden", arePluginsEnabled: true},
		{args: "dash-test", fixture: "plugin/dash-test1.golden", arePluginsEnabled: true},
		{args: "another_dash-test but-with two-args with dashes and-others_without them", fixture: "plugin/dash-test2.golden", arePluginsEnabled: true},
		{args: "cli command", fixture: "plugin/cli-commands.golden", regex: true, arePluginsEnabled: true},
		{args: "plugin list", fixture: "plugin/list.golden"},
	}

	resetConfiguration(s.T(), true)

	if runtime.GOOS != "windows" {
		for _, tt := range tests {
			tt.workflow = true
			s.runIntegrationTest(tt)
		}
	}
}
