// +build !windows

package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/stretchr/testify/require"
)

func (s *CLITestSuite) TestLocalHelpCommands() {
	var tests []CLITest
	tests = []CLITest{
		// These should all be equivalent
		{args: "local", fixture: "local-help1.golden"},
		{args: "help local", fixture: "local-help1.golden"},
		{args: "local --help", fixture: "local-help1.golden"},
		// Ideally, this would show subcommand help, but Cobra doesn't send "list" arg to the help command func.
		// So we just show the top level list of commands again. :(
		{args: "help local list", fixture: "local-help1.golden"},
		// But if we call it this way, we can see help for specific local subcommands.
		{args: "local list --help", fixture: "local-help2.golden"},
		// We only have help 2 command levels deep. "local list plugins --help" shows the same as "local list --help"
		{args: "local list plugins --help", fixture: "local-help2.golden"},
	}
	resetConfiguration(s.T(), "confluent")
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runConfluentTest(tt, serveMds(s.T(), kafkaAPIURL).URL)
	}
}

func (s *CLITestSuite) TestLocalVersion() {
	tests := []CLITest{
		{name: "5.3.1 community", args: "local --path %s version", fixture: "local-version1.golden"},
		{name: "5.3.1 enterprise", args: "local --path %s version", fixture: "local-version2.golden"},
		{name: "5.4.0 community", args: "local --path %s version", fixture: "local-version3.golden"},
		{name: "5.4.0 enterprise", args: "local --path %s version", fixture: "local-version4.golden"},
	}
	resetConfiguration(s.T(), "confluent")
	for _, tt := range tests {
		parts := strings.Split(tt.name, " ")
		version := parts[0]
		enterprise := parts[1] == "enterprise"
		path, err := makeFakeCPLocalInstall(version, enterprise)
		require.NoError(s.T(), err)
		//noinspection GoDeferInLoop
		defer os.RemoveAll(path) // clean up
		tt.args = fmt.Sprintf(tt.args, path)
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runConfluentTest(tt, serveMds(s.T(), kafkaAPIURL).URL)
	}
}

func makeFakeCPLocalInstall(version string, enterprise bool) (string, error) {
	path, err := ioutil.TempDir("", "confluent-"+version)
	if err != nil {
		return "", err
	}

	// setup to pass "validate" step
	srConf := fmt.Sprintf("%s/etc/schema-registry/", path)
	err = os.MkdirAll(srConf, os.ModePerm)
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s/connect-avro-distributed.properties", srConf), []byte(""), os.ModePerm)
	if err != nil {
		return "", err
	}

	// setup for version detection to correctly decide if its "Confluent Platform" or "Confluent Community Software"
	if enterprise {
		var filename string
		if version == "5.3.1" {
			filename = "kafka-connect-replicator-5.3.1.jar"
		} else {
			filename = "connect-replicator-5.4.0.jar"
		}
		replicatorDir := fmt.Sprintf("%s/share/java/kafka-connect-replicator/", path)
		err = os.MkdirAll(replicatorDir, os.ModePerm)
		if err != nil {
			return "", err
		}
		err = ioutil.WriteFile(fmt.Sprintf("%s/%s", replicatorDir, filename), []byte(""), os.ModePerm)
		if err != nil {
			return "", err
		}
	} else {
		filename := fmt.Sprintf("common-config-%s.jar", version)
		commonDir := fmt.Sprintf("%s/share/java/confluent-common/", path)
		err = os.MkdirAll(commonDir, os.ModePerm)
		if err != nil {
			return "", err
		}
		err = ioutil.WriteFile(fmt.Sprintf("%s/%s", commonDir, filename), []byte(""), os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	return path, nil
}
