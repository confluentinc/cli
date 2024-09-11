package test

import (
	"fmt"
	"io/fs"
	"os"
	"runtime"
	"strings"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v3/pkg/utils"
)

func (s *CLITestSuite) TestPlugin() {
	tests := []CLITest{
		{args: "plugin1 arg1", fixture: "plugin/plugin1.golden"},
		{args: "print args arg1 arg2 --meaningless-flag true arg3", fixture: "plugin/print-args.golden"},
		{args: "version", fixture: "version/version.golden", regex: true},
		{args: "kafka something kafkaesque", fixture: "plugin/partial-name-overlap.golden"},
		{args: "foo bar baz boo far foo bar baz --flag true", fixture: "plugin/long-plugin-name.golden"},
		{args: "can print to stderr --meaningless-flag false and stdout", fixture: "plugin/print-stderr.golden"},
		{args: "dash_test", fixture: "plugin/dash-test1.golden"},
		{args: "dash-test", fixture: "plugin/dash-test1.golden"},
		{args: "another_dash-test but-with two-args with dashes and-others_without them", fixture: "plugin/dash-test2.golden"},
		{args: "cli command", fixture: "version/version.golden", regex: true},
		{args: "plugin list", fixture: "plugin/list.golden"},
	}

	resetConfiguration(s.T(), true) // enable plugins

	path := "test/bin:test/fixtures/input/plugin:test/fixtures/input/plugin/test"
	if runtime.GOOS == "windows" {
		path = strings.ReplaceAll(path, ":", ";")
	}

	if runtime.GOOS != "windows" {
		for _, test := range tests {
			test.workflow = true
			test.env = []string{fmt.Sprintf("PATH=%s", path)}
			s.runIntegrationTest(test)
		}
	}
}

func (s *CLITestSuite) TestPluginUninstall() {
	req := require.New(s.T())
	filename := "test/fixtures/input/plugin/confluent-test-plugin-uninstall.sh"
	file, err := os.Create(filename)
	req.NoError(err)
	err = file.Chmod(fs.ModePerm)
	req.NoError(err)
	err = file.Close()
	req.NoError(err)
	defer func() {
		if utils.FileExists(filename) {
			err := os.Remove(filename)
			req.NoError(err)
		}
	}()

	tests := []CLITest{
		{args: "plugin uninstall confluent-test-plugin-uninstall", input: "y\n", fixture: "plugin/uninstall.golden"},
		{args: "plugin list", fixture: "plugin/list.golden"},
	}

	resetConfiguration(s.T(), true) // enable plugins

	path := "test/bin:test/fixtures/input/plugin:test/fixtures/input/plugin/test"
	if runtime.GOOS == "windows" {
		path = strings.ReplaceAll(path, ":", ";")
	}

	if runtime.GOOS != "windows" {
		for _, test := range tests {
			test.workflow = true
			test.env = []string{fmt.Sprintf("PATH=%s", path)}
			s.runIntegrationTest(test)
		}
	}
}

func (s *CLITestSuite) TestPlugin_Disabled() {
	tests := []CLITest{
		{args: "plugin1 arg1", fixture: "plugin/plugin1-disabled.golden", exitCode: 1},
		{args: "print args arg1 arg2 --meaningless-flag true arg3", fixture: "plugin/print-args-disabled.golden", exitCode: 1},
		{args: "plugin list", fixture: "plugin/list-disabled.golden", exitCode: 1},
	}

	resetConfiguration(s.T(), false) // disable plugins

	path := "test/fixtures/input/plugin:test/fixtures/input/plugin/test"
	if runtime.GOOS == "windows" {
		path = strings.ReplaceAll(path, ":", ";")
	}

	if runtime.GOOS != "windows" {
		for _, test := range tests {
			test.workflow = true
			test.env = []string{fmt.Sprintf("PATH=%s", path)}
			s.runIntegrationTest(test)
		}
	}
}
