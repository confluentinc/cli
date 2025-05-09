package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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

func (s *ConfluentHomeTestSuite) TestGetFile() {
	req := require.New(s.T())

	dir, err := s.ch.getRootDir()
	req.NoError(err)

	file, err := s.ch.GetFile(exampleDir, exampleFile)
	req.NoError(err)
	req.Equal(filepath.Join(dir, exampleDir, exampleFile), file)
}

func (s *ConfluentHomeTestSuite) TestHasFile() {
	req := require.New(s.T())

	dir, err := s.ch.getRootDir()
	req.NoError(err)

	has, err := s.ch.HasFile(exampleFile)
	req.NoError(err)
	req.False(has)

	_, err = os.Create(filepath.Join(dir, exampleFile))
	req.NoError(err)

	has, err = s.ch.HasFile(exampleFile)
	req.NoError(err)
	req.True(has)
}

func (s *ConfluentHomeTestSuite) TestIsConfluentPlatform() {
	req := require.New(s.T())

	file := "share/java/confluent-rebalancer/confluent-rebalancer-0.0.0.jar"
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

	version, err := s.ch.GetVersion(exampleService, true)
	req.NoError(err)
	req.Equal(exampleVersion, version)
}

func (s *ConfluentHomeTestSuite) TestGetVersionNoMatchError() {
	req := require.New(s.T())

	_, err := s.ch.GetVersion(exampleService, true)
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

	return os.WriteFile(path, []byte{}, 0644)
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
