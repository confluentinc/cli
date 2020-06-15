package local

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	exampleDir     = "dir"
	exampleFile    = "file"
	exampleService = "service"
)

var (
	ch       *ConfluentHomeManager
	dirCount = 0
)

func TestIsConfluentPlatform(t *testing.T) {
	req := require.New(t)

	setup(req)
	defer teardown()

	file := "share/java/confluent-control-center/control-center-0.0.0.jar"
	req.NoError(createTestConfluentFile(ch, file))

	isCP, err := ch.IsConfluentPlatform()
	req.NoError(err)
	req.True(isCP)
}

func TestIsNotConfluentPlatform(t *testing.T) {
	req := require.New(t)

	setup(req)
	defer teardown()

	isCP, err := ch.IsConfluentPlatform()
	req.NoError(err)
	req.False(isCP)
}

func TestFindFile(t *testing.T) {
	req := require.New(t)

	setup(req)
	defer teardown()

	req.NoError(createTestConfluentFile(ch, "file-0.0.0.txt"))

	matches, err := ch.FindFile("file-*.txt")
	req.NoError(err)
	req.Equal([]string{"file-0.0.0.txt"}, matches)
}

func setup(req *require.Assertions) {
	dir, err := createTestDir()
	req.NoError(err)
	req.NoError(os.Setenv("CONFLUENT_HOME", dir))
}

func teardown() {
	dir, _ := ch.getRootDir()
	os.RemoveAll(dir)
	os.Clearenv()
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

// Create an empty file inside of CONFLUENT_HOME
func createTestConfluentFile(ch *ConfluentHomeManager, file string) error {
	dir, err := ch.getRootDir()
	if err != nil {
		return err
	}

	path := filepath.Join(dir, file)
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}

	return ioutil.WriteFile(path, []byte{}, 0644)
}
