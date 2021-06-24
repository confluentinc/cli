package test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/stretchr/testify/require"
)

func (s *CLITestSuite) TestLocalLifecycle() {
	s.createCH([]string{
		"share/java/confluent-control-center/control-center-5.5.0.jar",
	})
	s.createCC()
	defer s.destroy()

	tests := []CLITest{
		{Args: "local destroy", Fixture: "local/destroy-error.golden", WantErrCode: 1},
		{Args: "local current", Fixture: "local/current.golden", Regex: true},
		{Args: "local destroy", Fixture: "local/destroy.golden", Regex: true},
	}

	for _, tt := range tests {
		tt.Workflow = true
		s.RunConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestLocalConfluentCommunitySoftware() {
	s.createCH([]string{
		"share/java/confluent-common/common-config-5.5.0.jar",
	})
	defer s.destroy()

	tests := []CLITest{
		{Args: "local services list", Fixture: "local/services-list-ccs.golden"},
		{Args: "local version", Fixture: "local/version-ccs.golden"},
	}

	for _, tt := range tests {
		s.RunConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestLocalVersion() {
	s.createCH([]string{
		"share/java/confluent-control-center/control-center-5.5.0.jar",
		"share/java/kafka-connect-replicator/connect-replicator-5.5.0.jar",
	})
	defer s.destroy()

	tests := []CLITest{
		{Args: "local version", Fixture: "local/version-cp.golden"},
	}

	for _, tt := range tests {
		s.RunConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestLocalServicesList() {
	s.createCH([]string{
		"share/java/confluent-control-center/control-center-5.5.0.jar",
	})
	defer s.destroy()

	tests := []CLITest{
		{Args: "local services list", Fixture: "local/services-list-cp.golden"},
	}

	for _, tt := range tests {
		s.RunConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestLocalServicesLifecycle() {
	s.createCH([]string{
		"share/java/confluent-control-center/control-center-5.5.0.jar",
	})
	defer s.destroy()

	tests := []CLITest{
		{Args: "local services status", Fixture: "local/services-status-all-stopped.golden", Regex: true},
		{Args: "local services stop", Fixture: "local/services-stop-already-stopped.golden", Regex: true},
		{Args: "local services top", Fixture: "local/services-top-no-services-running.golden", WantErrCode: 1},
	}

	for _, tt := range tests {
		s.RunConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestLocalZookeeperLifecycle() {
	s.createCH([]string{
		"share/java/kafka/zookeeper-5.5.0.jar",
	})
	defer s.destroy()

	tests := []CLITest{
		{Args: "local services zookeeper log", Fixture: "local/zookeeper-log-error.golden", WantErrCode: 1},
		{Args: "local services zookeeper status", Fixture: "local/zookeeper-status-stopped.golden", Regex: true},
		{Args: "local services zookeeper stop", Fixture: "local/zookeeper-stop-already-stopped.golden", Regex: true},
		{Args: "local services zookeeper top", Fixture: "local/zookeeper-top-stopped.golden"},
		{Args: "local services zookeeper version", Fixture: "local/zookeeper-version.golden"},
	}

	for _, tt := range tests {
		s.RunConfluentTest(tt)
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
