package plugin

import (
	"github.com/confluentinc/cli/internal/pkg/mock"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"regexp"
	"runtime"
	"testing"
)

func TestIsExec_Dir(t *testing.T) {
	f := &mock.FileInfo{ModeVal: fs.ModeDir}
	require.Equal(t, false, isExec(f))
}

func TestIsExec_Executable(t *testing.T) {
	f := &mock.FileInfo{ModeVal: fs.ModePerm}
	require.Equal(t, true, isExec(f))
}

func TestIsExec_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}
	f := &mock.FileInfo{NameVal: "hello.exe"}
	require.Equal(t, true, isExec(f))
}

func TestIsPluginFn(t *testing.T) {
	pluginMap := make(map[string][]string)
	re := regexp.MustCompile("confluent(-[a-z]+)+(\\.[a-z]+)?")
	f := isPluginFn(re, pluginMap)

	f("confluent-plugin1", &mock.FileInfo{ModeVal: fs.ModePerm}, nil)
	require.Equal(t, 1, len(pluginMap))

	f("onfluent-plugin1", &mock.FileInfo{ModeVal: fs.ModePerm}, nil)
	require.Equal(t, 1, len(pluginMap))

	f("confluent-", &mock.FileInfo{ModeVal: fs.ModePerm}, nil)
	require.Equal(t, 1, len(pluginMap))

	f("confluent", &mock.FileInfo{ModeVal: fs.ModePerm}, nil)
	require.Equal(t, 1, len(pluginMap))

	f("confluent-foo-bar-baz", &mock.FileInfo{ModeVal: fs.ModePerm}, nil)
	require.Equal(t, 2, len(pluginMap))

	f("confluent-foo-bar", &mock.FileInfo{ModeVal: fs.ModeDir}, nil)
	require.Equal(t, 2, len(pluginMap))
}

func TestSearchPath(t *testing.T) {

	root, err := os.MkdirTemp(os.TempDir(), "plugin_test")
	defer os.RemoveAll(root)
	require.NoError(t, err)
	file, err := os.CreateTemp(root, "confluent-plugin.sh")
	file.Chmod(fs.ModePerm)
	require.NoError(t, err)
	path := os.Getenv("PATH")
	os.Setenv("PATH", root)
	pluginMap, err := SearchPath()
	require.NoError(t, err)
	require.Equal(t, 1, len(pluginMap))
	os.Setenv("PATH", path)
}
