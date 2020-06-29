package test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/stretchr/testify/require"
)

func (s *CLITestSuite) TestLocalLifecycle() {
	s.createCH([]string{
		"share/java/confluent-control-center/control-center-0.0.0.jar",
	})
	s.createCC()
	defer s.destroy()

	tests := []CLITest{
		{args: "local-v2 destroy", fixture: "local/destroy-error.golden", wantErrCode: 1},
		{args: "local-v2 current", fixture: "local/current.golden", regex: true},
		{args: "local-v2 destroy", fixture: "local/destroy.golden", regex: true},
	}

	for _, test := range tests {
		s.runConfluentTest(test, serveMds(s.T()).URL)
	}
}

func (s *CLITestSuite) TestLocalConfluentCommunitySoftware() {
	s.createCH([]string{
		"share/java/confluent-common/common-config-0.0.0.jar",
	})
	defer s.destroy()

	tests := []CLITest{
		{args: "local-v2 version", fixture: "local/version-ccs.golden"},
		{args: "local-v2 services list", fixture: "local/services-list-ccs.golden"},
	}

	for _, test := range tests {
		s.runConfluentTest(test, serveMds(s.T()).URL)
	}
}

// TODO: Change name to TestLocalVersion after deleting old code
func (s *CLITestSuite) TestLocalV2Version() {
	s.createCH([]string{
		"share/java/confluent-control-center/control-center-0.0.0.jar",
		"share/java/kafka-connect-replicator/connect-replicator-0.0.0.jar",
	})
	defer s.destroy()

	test := CLITest{args: "local-v2 version", fixture: "local/version-cp.golden"}
	s.runConfluentTest(test, serveMds(s.T()).URL)
}

func (s *CLITestSuite) TestLocalServicesList() {
	s.createCH([]string{
		"share/java/confluent-control-center/control-center-0.0.0.jar",
	})
	defer s.destroy()

	test := CLITest{args: "local-v2 services list", fixture: "local/services-list-cp.golden"}
	s.runConfluentTest(test, serveMds(s.T()).URL)
}

func (s *CLITestSuite) TestLocalServicesLifecycle() {
	s.createCH([]string{
		"share/java/confluent-control-center/control-center-0.0.0.jar",
	})
	defer s.destroy()

	tests := []CLITest{
		{args: "local-v2 services status", fixture: "local/services-status-all-stopped.golden", regex: true},
		{args: "local-v2 services stop", fixture: "local/services-stop-already-stopped.golden", regex: true},
		{args: "local-v2 services top", fixture: "local/services-top-no-services-running.golden", wantErrCode: 1},
	}

	for _, test := range tests {
		s.runConfluentTest(test, serveMds(s.T()).URL)
	}
}

func (s *CLITestSuite) TestLocalServiceLifecycle() {
	s.createCH([]string{
		"share/java/kafka/zookeeper-0.0.0.jar",
	})
	defer s.destroy()

	tests := []CLITest{
		{args: "local-v2 services zookeeper log", fixture: "local/zookeeper-log-error.golden", wantErrCode: 1},
		{args: "local-v2 services zookeeper status", fixture: "local/zookeeper-status-stopped.golden", regex: true},
		{args: "local-v2 services zookeeper stop", fixture: "local/zookeeper-stop-already-stopped.golden", regex: true},
		{args: "local-v2 services zookeeper top", fixture: "local/zookeeper-top-stopped.golden"},
		{args: "local-v2 services zookeeper version", fixture: "local/zookeeper-version.golden"},
	}

	for _, test := range tests {
		s.runConfluentTest(test, serveMds(s.T()).URL)
	}
}

func (s *CLITestSuite) createCC() {
	req := require.New(s.T())

	dir := filepath.Join(os.TempDir(), "confluent-int-test", "cc")
	req.NoError(os.Setenv("CONFLUENT_CURRENT", dir))
}

func (s *CLITestSuite) createCH(files []string) {
	req := require.New(s.T())

	dir := filepath.Join(os.TempDir(), "confluent-int-test", "ch")
	req.NoError(os.Setenv("CONFLUENT_HOME", dir))

	for _, file := range files {
		path := filepath.Join(dir, file)

		dir := filepath.Dir(path)
		req.NoError(os.MkdirAll(dir, 0777))

		req.NoError(ioutil.WriteFile(path, []byte{}, 0644))
	}
}

func (s *CLITestSuite) destroy() {
	req := require.New(s.T())

	req.NoError(os.Setenv("CONFLUENT_HOME", ""))
	req.NoError(os.Setenv("CONFLUENT_CURRENT", ""))
	dir := filepath.Join(os.TempDir(), "confluent-int-test")
	req.NoError(os.RemoveAll(dir))
}
