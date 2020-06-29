package local

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSavedCurrentDir(t *testing.T) {
	req := require.New(t)

	cc := NewConfluentCurrentManager()
	cc.currentDir = exampleDir

	dir, err := cc.GetCurrentDir()
	req.NoError(err)
	req.Equal(exampleDir, dir)
}

func TestCreateAndTrackCurrentDir(t *testing.T) {
	req := require.New(t)

	dir, err := createTestDir()
	req.NoError(err)
	defer os.RemoveAll(dir)

	req.NoError(os.Setenv("CONFLUENT_CURRENT", dir))
	defer os.Clearenv()

	cc := NewConfluentCurrentManager()

	currentDir, err := cc.GetCurrentDir()
	req.NoError(err)

	req.DirExists(currentDir)
	req.FileExists(cc.trackingFile)
	data, err := ioutil.ReadFile(cc.trackingFile)
	req.NoError(err)
	req.Equal(currentDir, strings.TrimSuffix(string(data), "\n"))
}

func TestGetCurrentDirFromTrackingFile(t *testing.T) {
	req := require.New(t)

	dir, err := createTestDir()
	req.NoError(err)
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, exampleFile)
	req.NoError(ioutil.WriteFile(file, []byte(exampleDir), 0644))

	cc := NewConfluentCurrentManager()
	cc.trackingFile = file

	currentDir, err := cc.GetCurrentDir()
	req.NoError(err)
	req.Equal(exampleDir, currentDir)
}

func TestGetServiceDir(t *testing.T) {
	req := require.New(t)

	dir, err := createTestDir()
	req.NoError(err)
	defer os.RemoveAll(dir)

	cc := NewConfluentCurrentManager()
	cc.currentDir = dir

	serviceDir, err := cc.getServiceDir(exampleService)
	req.NoError(err)
	req.DirExists(serviceDir)
}

func TestGetDataDir(t *testing.T) {
	req := require.New(t)

	dir, err := createTestDir()
	req.NoError(err)
	defer os.RemoveAll(dir)

	cc := NewConfluentCurrentManager()
	cc.currentDir = dir

	_, err = cc.getServiceDir(exampleService)
	req.NoError(err)

	dataDir, err := cc.GetDataDir(exampleService)
	req.NoError(err)
	req.DirExists(dataDir)
}

func TestGetDataDirKSQL(t *testing.T) {
	req := require.New(t)

	dir, err := createTestDir()
	req.NoError(err)
	defer os.RemoveAll(dir)

	cc := NewConfluentCurrentManager()
	cc.currentDir = dir

	_, err = cc.getServiceDir("ksql-server")
	req.NoError(err)

	dataDir, err := cc.GetDataDir("ksql-server")
	req.NoError(err)
	req.DirExists(dataDir)
	req.Contains(dataDir, "kafka-streams")
}

func TestGetSavedPidFile(t *testing.T) {
	req := require.New(t)

	cc := NewConfluentCurrentManager()
	cc.pidFiles[exampleService] = exampleFile

	file, err := cc.GetPidFile(exampleService)
	req.NoError(err)
	req.Equal(exampleFile, file)
}

func TestSetAndGetPid(t *testing.T) {
	req := require.New(t)

	dir, err := createTestDir()
	req.NoError(err)
	defer os.RemoveAll(dir)

	cc := NewConfluentCurrentManager()
	cc.currentDir = dir

	req.NoError(cc.WritePid(exampleService, 1))
	pid, err := cc.ReadPid(exampleService)
	req.NoError(err)
	req.Equal(1, pid)
}

func TestGetDefaultRootDir(t *testing.T) {
	req := require.New(t)

	cc := NewConfluentCurrentManager()
	req.Equal(os.TempDir(), cc.getRootDir())
}

func TestRootDirFromEnv(t *testing.T) {
	req := require.New(t)

	cc := NewConfluentCurrentManager()
	req.NoError(os.Setenv("CONFLUENT_CURRENT", exampleDir))
	defer os.Clearenv()

	req.Equal(exampleDir, cc.getRootDir())
}

func TestGetSavedTrackingFile(t *testing.T) {
	req := require.New(t)

	cc := NewConfluentCurrentManager()
	cc.trackingFile = exampleFile
	req.Equal(exampleFile, cc.getTrackingFile())
}

func TestGetTrackingFile(t *testing.T) {
	req := require.New(t)

	cc := NewConfluentCurrentManager()
	req.NoError(os.Setenv("CONFLUENT_CURRENT", exampleDir))
	defer os.Clearenv()

	path := filepath.Join(exampleDir, "confluent.current")
	req.Equal(path, cc.getTrackingFile())
}

func TestGetServiceFile(t *testing.T) {
	req := require.New(t)

	dir, err := createTestDir()
	req.NoError(err)
	defer os.RemoveAll(dir)

	cc := NewConfluentCurrentManager()
	cc.currentDir = dir

	file, err := cc.getServiceFile(exampleService, exampleFile)
	req.NoError(err)
	req.Equal(filepath.Join(dir, exampleService, exampleFile), file)
}

func TestGetRandomChildDir(t *testing.T) {
	req := require.New(t)

	parentDir, err := createTestDir()
	req.NoError(err)
	defer os.RemoveAll(parentDir)

	childDir := getRandomChildDir(parentDir)
	req.Regexp(regexp.MustCompile(`confluent.\d{6}`), childDir)
}
