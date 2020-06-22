package local

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/stretchr/testify/require"
)

const (
	exampleDir     = "dir"
	exampleFile    = "file"
	exampleService = "kafka"
	exampleVersion = "0.0.0"
)

var dirCount = 0

type ConfluentHomeTestSuite struct {
	suite.Suite
	ch *ConfluentHomeManager
}

func TestConfluentHomeTestSuite(t *testing.T) {
	suite.Run(t, new(ConfluentHomeTestSuite))
}

func (s *ConfluentHomeTestSuite) SetupTest() {
	s.ch = NewConfluentHomeManager()
	dir, _ := createTestDir()
	os.Setenv("CONFLUENT_HOME", dir)
}

func (s *ConfluentHomeTestSuite) TearDownTest() {
	dir, _ := s.ch.getRootDir()
	os.RemoveAll(dir)
	os.Clearenv()
}

func (s *ConfluentHomeTestSuite) TestIsConfluentPlatform() {
	req := require.New(s.T())

	file := "share/java/confluent-control-center/control-center-0.0.0.jar"
	req.NoError(s.createTestConfluentFile(file))

	isCP, err := s.ch.IsConfluentPlatform()
	req.NoError(err)
	req.True(isCP)
}

func (s *ConfluentHomeTestSuite) TestIsNotConfluentPlatform() {
	req := require.New(s.T())

	isCP, err := s.ch.IsConfluentPlatform()
	req.NoError(err)
	req.False(isCP)
}

func (s *ConfluentHomeTestSuite) TestFindFile() {
	req := require.New(s.T())

	req.NoError(s.createTestConfluentFile("file-0.0.0.txt"))

	matches, err := s.ch.FindFile("file-*.txt")
	req.NoError(err)
	req.Equal([]string{"file-0.0.0.txt"}, matches)
}

func (s *ConfluentHomeTestSuite) TestGetVersion() {
	req := require.New(s.T())

	file := strings.ReplaceAll(versionFiles[exampleService], "*", exampleVersion)
	req.NoError(s.createTestConfluentFile(file))

	version, err := s.ch.GetVersion(exampleService)
	req.NoError(err)
	req.Equal(exampleVersion, version)
}

func (s *ConfluentHomeTestSuite) TestGetVersionNoMatchError() {
	req := require.New(s.T())

	_, err := s.ch.GetVersion(exampleService)
	req.Error(err)
}

// Create an empty file inside of CONFLUENT_HOME
func (s *ConfluentHomeTestSuite) createTestConfluentFile(file string) error {
	dir, err := s.ch.getRootDir()
	if err != nil {
		return err
	}

	path := filepath.Join(dir, file)
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}

	return ioutil.WriteFile(path, []byte{}, 0644)
}

// Directories must have unique names to satisfy Windows tests
func createTestDir() (string, error) {
	dir := fmt.Sprintf("confluent.test-dir.%06d", dirCount)
	dirCount++

	path := filepath.Join(os.TempDir(), dir)
	if err := os.MkdirAll(path, 0777); err != nil {
		return "", err
	}

	return path, nil
}
