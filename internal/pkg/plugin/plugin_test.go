package plugin

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/mock"
)

func TestIsExec_Dir(t *testing.T) {
	f := &mock.FileInfo{ModeVal: fs.ModeDir}
	require.False(t, isExecutable(f))
}

func TestIsExec_Executable(t *testing.T) {
	if runtime.GOOS == "windows" {
		require.True(t, isExecutableWindows("hello.exe"))
	} else {
		require.True(t, isExecutable(&mock.FileInfo{ModeVal: fs.ModePerm}))
	}
}

type pluginWalkInfo struct {
	name       string
	fileMode   fs.FileMode
	expectSize int
}

func TestPluginWalkFn(t *testing.T) {
	re := regexp.MustCompile(`confluent(-[a-z]+)+(\.[a-z]+)?`)
	if runtime.GOOS == "windows" {
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
			assert.Equal(t, test.expectSize, len(pluginMap))
		}
	} else {
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
			assert.Equal(t, test.expectSize, len(pluginMap))
		}
	}
}

func TestSearchPath(t *testing.T) {
	root, err := os.MkdirTemp(os.TempDir(), "plugin_test")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(root)
	}()
	var fileName, pluginName string
	pattern := "confluent-plugin"
	if runtime.GOOS == "windows" {
		pattern += "*.exe"
	}
	file, err := os.CreateTemp(root, pattern)
	require.NoError(t, err)
	fileName = filepath.Base(file.Name())
	if runtime.GOOS == "windows" {
		pluginName = fileName[:strings.LastIndex(fileName, ".")]
	} else {
		pluginName = fileName
		err = file.Chmod(fs.ModePerm)
		require.NoError(t, err)
	}
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
