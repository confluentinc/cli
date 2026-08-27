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

	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/mock"
	pversion "github.com/confluentinc/cli/v4/pkg/version"
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

func TestNameFromEntry(t *testing.T) {
	if runtime.GOOS == "windows" {
		tests := []pluginWalkInfo{
			{"confluent-plugin1.exe", fs.ModePerm, "confluent-plugin1"},
			{"confluent-foo-bar-baz", fs.ModePerm, ""},
			{"confluent-foo-bar.bat", fs.ModePerm, "confluent-foo-bar"},
		}

		for _, test := range tests {
			name := nameFromEntry(&mock.FileInfo{
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
			name := nameFromEntry(&mock.FileInfo{
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

	pluginMap := SearchPath(&config.Config{})
	pluginPaths, ok := pluginMap[pluginName]
	require.True(t, ok)
	require.Equal(t, fileName, filepath.Base(pluginPaths[0]))
}

// SearchPath also scans the channel-scoped state directory ($HOME/<StateDirName>/plugins), not just
// $PATH; a regression there would make installed plugins undiscoverable while TestSearchPath (which
// only exercises $PATH) still passed.
func TestSearchPath_ChannelStateDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Cleanup(func() { pversion.SetProcessChannel(pversion.Dev) })
	pversion.SetProcessChannel(pversion.Dev)

	// An empty $PATH forces discovery through the state directory alone.
	t.Setenv("PATH", t.TempDir())

	pluginDir := filepath.Join(home, ".confluent-dev", "plugins")
	require.NoError(t, os.MkdirAll(pluginDir, 0700))
	name := writeFakePlugin(t, pluginDir, "confluent-foo")

	pluginPaths, ok := SearchPath(&config.Config{})["confluent-foo"]

	require.True(t, ok, "a plugin under the channel state directory must be discovered")
	require.Equal(t, name, filepath.Base(pluginPaths[0]))
}

// When the home directory is unresolvable, config.StateDir fails and SearchPath degrades to $PATH
// alone. The failure mode this guards against is the old behavior that joined a relative
// ".confluent/plugins" (empty home + Join), which resolved against the working directory and would
// run whatever confluent-* executables happened to sit there.
func TestSearchPath_StateDirError(t *testing.T) {
	// os.UserHomeDir reads HOME on Unix and USERPROFILE on Windows; clearing both makes it error,
	// which is what drives config.StateDir to fail.
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	// Plant a plugin in a working-directory-relative ".confluent/plugins": a regressed degrade path
	// would scan it, the correct one must not.
	t.Chdir(t.TempDir())
	relativePluginDir := filepath.Join(".confluent", "plugins")
	require.NoError(t, os.MkdirAll(relativePluginDir, 0700))
	writeFakePlugin(t, relativePluginDir, "confluent-cwd")

	// A plugin on $PATH must still be discovered with no reachable state directory.
	pathDir := t.TempDir()
	name := writeFakePlugin(t, pathDir, "confluent-foo")
	t.Setenv("PATH", pathDir)

	plugins := SearchPath(&config.Config{})

	pluginPaths, ok := plugins["confluent-foo"]
	require.True(t, ok, "plugins on $PATH must still be discovered when the state directory is unavailable")
	require.Equal(t, name, filepath.Base(pluginPaths[0]))
	require.NotContains(t, plugins, "confluent-cwd", "a working-directory-relative plugins directory must not be scanned")
}

// writeFakePlugin creates an empty executable for a plugin named base (with the Windows .exe suffix
// where applicable) in dir, and returns the file name written.
func writeFakePlugin(t *testing.T, dir, base string) string {
	if runtime.GOOS == "windows" {
		base += ".exe"
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, base), nil, fs.ModePerm))
	return base
}

func TestVersionRegex(t *testing.T) {
	// Go
	goInstaller := &GoPluginInstaller{}
	require.True(t, goInstaller.IsVersion("go1.20"))
	require.True(t, goInstaller.IsVersion("go1.19.6"))
	require.False(t, goInstaller.IsVersion("1.19.6"))
	require.False(t, goInstaller.IsVersion("go1.19.0"))
	require.False(t, goInstaller.IsVersion("go"))
	require.False(t, goInstaller.IsVersion("version"))

	// Python
	pythonInstaller := &PythonPluginInstaller{}
	require.True(t, pythonInstaller.IsVersion("3.11.4"))
	require.True(t, pythonInstaller.IsVersion("3.11.0"))
	require.True(t, pythonInstaller.IsVersion("2.7.0"))
	require.False(t, pythonInstaller.IsVersion("Python"))

	// Bash
	bashInstaller := &BashPluginInstaller{}
	require.True(t, bashInstaller.IsVersion("3.2.57(1)-release"))
	require.False(t, bashInstaller.IsVersion("3.2.57(1)"))
	require.False(t, bashInstaller.IsVersion("3.2.57"))
	require.False(t, bashInstaller.IsVersion("bash"))
	require.False(t, bashInstaller.IsVersion("Inc."))
}

func TestToCommandName(t *testing.T) {
	require.Equal(t, "confluent login headless-sso", ToCommandName("confluent-login-headless_sso"))
}
