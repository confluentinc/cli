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
		{args: "local destroy", fixture: "local/destroy-error.golden", wantErrCode: 1},
		{args: "local current", fixture: "local/current.golden", regex: true},
		{args: "local destroy", fixture: "local/destroy.golden", regex: true},
	}

	for _, tt := range tests {
		tt.workflow = true
		s.runConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestLocalConfluentCommunitySoftware() {
	s.createCH([]string{
		"share/java/confluent-common/common-config-5.5.0.jar",
	})
	defer s.destroy()

	tests := []CLITest{
		{args: "local services list", fixture: "local/services-list-ccs.golden"},
		{args: "local version", fixture: "local/version-ccs.golden"},
	}

	for _, tt := range tests {
		s.runConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestLocalVersion() {
	s.createCH([]string{
		"share/java/confluent-control-center/control-center-5.5.0.jar",
		"share/java/kafka-connect-replicator/connect-replicator-5.5.0.jar",
	})
	defer s.destroy()

	tests := []CLITest{
		{args: "local version", fixture: "local/version-cp.golden"},
	}

	for _, tt := range tests {
		s.runConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestLocalServicesList() {
	s.createCH([]string{
		"share/java/confluent-control-center/control-center-5.5.0.jar",
	})
	defer s.destroy()

	tests := []CLITest{
		{args: "local services list", fixture: "local/services-list-cp.golden"},
	}

	for _, tt := range tests {
		s.runConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestLocalServicesLifecycle() {
	s.createCH([]string{
		"share/java/confluent-control-center/control-center-5.5.0.jar",
	})
	defer s.destroy()

	tests := []CLITest{
		{args: "local services status", fixture: "local/services-status-all-stopped.golden", regex: true},
		{args: "local services stop", fixture: "local/services-stop-already-stopped.golden", regex: true},
		{args: "local services top", fixture: "local/services-top-no-services-running.golden", wantErrCode: 1},
	}

	for _, tt := range tests {
		s.runConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestLocalZookeeperLifecycle() {
	s.createCH([]string{
		"share/java/kafka/zookeeper-5.5.0.jar",
	})
	defer s.destroy()

	tests := []CLITest{
		{args: "local services zookeeper log", fixture: "local/zookeeper-log-error.golden", wantErrCode: 1},
		{args: "local services zookeeper status", fixture: "local/zookeeper-status-stopped.golden", regex: true},
		{args: "local services zookeeper stop", fixture: "local/zookeeper-stop-already-stopped.golden", regex: true},
		{args: "local services zookeeper top", fixture: "local/zookeeper-top-stopped.golden"},
		{args: "local services zookeeper version", fixture: "local/zookeeper-version.golden"},
	}

	for _, tt := range tests {
		s.runConfluentTest(tt)
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
