package plugin

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/config"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/mock"
)

func TestIsExec_Dir(t *testing.T) {
	f := &mock.FileInfo{ModeVal: fs.ModeDir}
	require.False(t, isExecutable(f))
}

func TestIsExec_Executable(t *testing.T) {
	if runtime.GOOS == "windows" {
		assert.False(t, isExecutable(&mock.FileInfo{NameVal: "hello.nonexe"}))
		assert.True(t, isExecutable(&mock.FileInfo{NameVal: "hello.exe"}))
	} else {
		assert.False(t, isExecutable(&mock.FileInfo{ModeVal: fs.ModeDir}))
		assert.True(t, isExecutable(&mock.FileInfo{ModeVal: fs.ModePerm}))
	}
}

type pluginWalkInfo struct {
	path     string
	fileMode fs.FileMode
	name     string
}

func TestPluginFromEntry(t *testing.T) {
	if runtime.GOOS == "windows" {
		tests := []pluginWalkInfo{
			{"confluent-plugin1.exe", fs.ModePerm, "confluent-plugin1"},
			{"confluent-foo-bar-baz", fs.ModePerm, ""},
			{"confluent-foo-bar.bat", fs.ModePerm, "confluent-foo-bar"},
		}

		for _, test := range tests {
			name := PluginFromEntry(&mock.FileInfo{
				NameVal: test.path,
				ModeVal: test.fileMode,
			})
			assert.Equal(t, test.name, name)
		}
	} else {
		tests := []pluginWalkInfo{
			{"confluent-plugin1", fs.ModePerm, "confluent-plugin1"},
			{"onfluent-plugin1", fs.ModePerm, ""},
			{"confluent-", fs.ModePerm, ""},
			{"confluent", fs.ModePerm, ""},
			{"confluent-foo-bar-baz.sh", fs.ModePerm, "confluent-foo-bar-baz"},
			{"confluent-foo-bar", fs.ModeDir, ""},
		}

		for _, test := range tests {
			name := PluginFromEntry(&mock.FileInfo{
				NameVal: test.path,
				ModeVal: test.fileMode,
			})
			assert.Equal(t, test.name, name)
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

	t.Setenv("PATH", root)

	pluginMap := SearchPath(&v1.Config{BaseConfig: new(config.BaseConfig)})
	pluginPaths, ok := pluginMap[pluginName]
	require.True(t, ok)
	require.Equal(t, fileName, filepath.Base(pluginPaths[0]))
}
