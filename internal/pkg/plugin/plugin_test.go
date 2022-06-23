package plugin

import (
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/mock"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
)

func TestIsExec_Dir(t *testing.T) {
	f := &mock.FileInfo{ModeVal: fs.ModeDir}
	require.Equal(t, false, isExec(f))
}

func TestIsExec_Executable(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	f := &mock.FileInfo{ModeVal: fs.ModePerm}
	require.Equal(t, true, isExec(f))
}

func TestIsExec_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}
	require.Equal(t, true, isExecWindows("hello.exe"))
}

func TestPluginWalkFn(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	pluginMap := make(map[string][]string)
	re := regexp.MustCompile(`confluent(-[a-z]+)+(\.[a-z]+)?`)
	f := pluginWalkFn(re, pluginMap)

	err := f("confluent-plugin1", &mock.FileInfo{ModeVal: fs.ModePerm}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(pluginMap))

	err = f("onfluent-plugin1", &mock.FileInfo{ModeVal: fs.ModePerm}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(pluginMap))

	err = f("confluent-", &mock.FileInfo{ModeVal: fs.ModePerm}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(pluginMap))

	err = f("confluent", &mock.FileInfo{ModeVal: fs.ModePerm}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(pluginMap))

	err = f("confluent-foo-bar-baz.sh", &mock.FileInfo{ModeVal: fs.ModePerm}, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(pluginMap))

	err = f("confluent-foo-bar", &mock.FileInfo{ModeVal: fs.ModeDir}, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(pluginMap))
}

func TestPluginWalkFn_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}
	pluginMap := make(map[string][]string)
	re := regexp.MustCompile(`confluent(-[a-z]+)+(\.[a-z]+)?`)
	f := pluginWalkFn(re, pluginMap)

	err := f("confluent-plugin1.exe", &mock.FileInfo{NameVal: "confluent-plugin1.exe", ModeVal: fs.ModePerm}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(pluginMap))

	err = f("confluent-foo-bar-baz", &mock.FileInfo{NameVal: "confluent-foo-bar-baz", ModeVal: fs.ModePerm}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(pluginMap))

	err = f("confluent-foo-bar.bat", &mock.FileInfo{NameVal: "confluent-foo-bar.bat", ModeVal: fs.ModePerm}, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(pluginMap))
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
	fmt.Println(pluginMap)
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
	fmt.Println(fileName)
	fmt.Println(pluginMap)
	require.NoError(t, err)
	pluginPaths, ok := pluginMap[fileName]
	require.True(t, ok)
	require.Equal(t, fileName, filepath.Base(pluginPaths[0]))
}
