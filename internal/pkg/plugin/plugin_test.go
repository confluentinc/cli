package plugin

import (
	"github.com/confluentinc/cli/internal/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestIsExec_Dir(t *testing.T) {
	f := &mock.FileInfo{ModeVal: fs.ModeDir}
	require.False(t, isExecutable(f))
}

func TestIsExec_Executable(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	f := &mock.FileInfo{ModeVal: fs.ModePerm}
	require.True(t, isExecutable(f))
}

func TestIsExec_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}
	require.True(t, isExecutableWindows("hello.exe"))
}

type pluginWalkInfo struct {
	name       string
	fileMode   fs.FileMode
	expectSize int
}

func TestPluginWalkFn(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	re := regexp.MustCompile(`confluent(-[a-z]+)+(\.[a-z]+)?`)
	tests := []pluginWalkInfo{
		{"confluent-plugin1", fs.ModePerm, 1},
		{"onfluent-plugin1", fs.ModePerm, 0},
		{"confluent-", fs.ModePerm, 0},
		{"confluent", fs.ModePerm, 0},
		{"confluent-foo-bar-baz.sh", fs.ModePerm, 1},
		{"confluent-foo-bar", fs.ModeDir, 0},
	}

	for _, test := range tests {
		pluginMap := make(map[string][]string)
		f := pluginWalkFn(re, pluginMap)
		err := f(test.name, &mock.FileInfo{ModeVal: test.fileMode}, nil)
		assert.NoError(t, err)
		assert.True(t, len(pluginMap) == test.expectSize)
	}
}

func TestPluginWalkFn_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}
	re := regexp.MustCompile(`confluent(-[a-z]+)+(\.[a-z]+)?`)
	tests := []pluginWalkInfo{
		{"confluent-plugin1.exe", fs.ModePerm, 1},
		{"confluent-foo-bar-baz", fs.ModePerm, 0},
		{"confluent-foo-bar.bat", fs.ModePerm, 1},
	}

	for _, test := range tests {
		pluginMap := make(map[string][]string)
		f := pluginWalkFn(re, pluginMap)
		err := f(test.name, &mock.FileInfo{NameVal: test.name, ModeVal: test.fileMode}, nil)
		assert.NoError(t, err)
		assert.True(t, len(pluginMap) == test.expectSize)
	}
}

func TestSearchPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	root, err := os.MkdirTemp(os.TempDir(), "plugin_test")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(root)
		require.NoError(t, err)
	}()

	file, err := os.CreateTemp(root, "confluent-plugin")
	require.NoError(t, err)
	fileName := filepath.Base(file.Name())

	err = file.Chmod(fs.ModePerm)
	require.NoError(t, err)

	path := os.Getenv("PATH")
	err = os.Setenv("PATH", root)
	require.NoError(t, err)
	defer func() {
		err := os.Setenv("PATH", path)
		require.NoError(t, err)
	}()

	pluginMap, err := SearchPath()
	require.NoError(t, err)
	pluginPaths, ok := pluginMap[fileName]
	require.True(t, ok)
	require.Equal(t, fileName, filepath.Base(pluginPaths[0]))
}

func TestSearchPath_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}
	root, err := os.MkdirTemp(os.TempDir(), "plugin_test")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(root)
	}()

	file, err := os.CreateTemp(root, "confluent-plugin*.exe")
	require.NoError(t, err)
	fileName := filepath.Base(file.Name())
	pluginName := fileName[:strings.LastIndex(fileName, ".")]

	path := os.Getenv("PATH")
	err = os.Setenv("PATH", root)
	require.NoError(t, err)
	defer func() {
		err := os.Setenv("PATH", path)
		require.NoError(t, err)
	}()

	pluginMap, err := SearchPath()
	require.NoError(t, err)
	pluginPaths, ok := pluginMap[pluginName]
	require.True(t, ok)
	require.Equal(t, fileName, filepath.Base(pluginPaths[0]))
}
