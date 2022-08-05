package plugin

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/utils"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

type pluginInfo struct {
	args []string
	idx  int
	name string
}

func SearchPath() (map[string][]string, error) {
	pluginMap := make(map[string][]string)
	re := regexp.MustCompile(`^confluent(-[a-z][0-9_a-z]*)+(\.[a-z]+)?$`)
	delimiter := ":"
	if runtime.GOOS == "windows" {
		delimiter = ";"
	}

	for _, dir := range strings.Split(os.Getenv("PATH"), delimiter) {
		dirName, err := homedir.Expand(dir)
		if err != nil {
			return nil, err
		}
		if err := filepath.WalkDir(dirName, pluginWalkFn(re, pluginMap)); err != nil {
			return nil, err
		}
	}
	return pluginMap, nil
}

func pluginWalkFn(re *regexp.Regexp, pluginMap map[string][]string) func(string, fs.DirEntry, error) error {
	return func(path string, entry fs.DirEntry, _ error) error {
		pluginName := filepath.Base(path)
		if re.MatchString(pluginName) && isExecutable(entry) {
			if strings.Contains(pluginName, ".") {
				pluginName = strings.TrimSuffix(pluginName, filepath.Ext(pluginName))
			}
			pluginMap[pluginName] = append(pluginMap[pluginName], path)
		}
		return nil
	}
}

func isExecutable(entry fs.DirEntry) bool {
	if runtime.GOOS == "windows" {
		return isExecutableWindows(entry.Name())
	}
	fileInfo, err := entry.Info()
	if err != nil {
		return false
	}
	return !fileInfo.Mode().IsDir() && fileInfo.Mode()&0111 != 0
}

func isExecutableWindows(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return utils.Contains([]string{".bat", ".cmd", ".com", ".exe", ".ps1"}, ext)
}

// FindPlugin determines if the arguments passed in are meant for a plugin
func FindPlugin(cmd *cobra.Command, args []string) (*pluginInfo, error) {
	pluginMap, err := SearchPath()
	if err != nil {
		return nil, err
	}

	plugin := newPluginInfo(args)

	for len(plugin.name) > len(pversion.CLIName) {
		if pluginPathList, ok := pluginMap[plugin.name]; ok {
			if cmd, _, _ := cmd.Find(args); strings.ReplaceAll(cmd.CommandPath(), " ", "-") == plugin.name {
				log.CliLogger.Warnf("[WARN] User plugin %s is ignored because its command line invocation matches existing CLI command `%s`.\n", pluginPathList[0], cmd.CommandPath())
				break
			}
			plugin.args = append([]string{pluginPathList[0]}, plugin.args...)
			return plugin, nil
		}
		plugin.args = append([]string{args[plugin.idx-1]}, plugin.args...)
		plugin.idx--
		plugin.name = plugin.name[:strings.LastIndex(plugin.name, "-")]
	}
	return nil, err
}

// newPluginInfo initializes a pluginInfo struct from command line arguments
func newPluginInfo(args []string) *pluginInfo {
	infoArgs := make([]string, 0, len(args))
	name := []string{pversion.CLIName}
	for i, arg := range args {
		if strings.HasPrefix(arg, "--") {
			infoArgs = args[i:]
			break
		}
		arg = strings.ReplaceAll(arg, "-", "_")
		name = append(name, arg)
	}
	return &pluginInfo{
		args: infoArgs,
		name: strings.Join(name, "-"),
		idx:  len(name) - 1,
	}
}

func SetupExecPlugin(plugin *pluginInfo) error {
	pluginPath := plugin.args[0]
	if strings.HasSuffix(pluginPath, ".sh") {
		dat, err := os.ReadFile(pluginPath)
		if err != nil {
			return err
		}
		if string(dat[:2]) != "#!" {
			shell := os.Getenv("SHELL")
			if shell == "" {
				shell = "/bin/bash"
			}
			shebang := []byte("#!" + shell + "\n")
			temp, err := os.CreateTemp("", plugin.name)
			if err != nil {
				return err
			}
			defer func() {
				_ = os.Remove(temp.Name())
			}()
			if _, err := temp.Write(append(shebang, dat...)); err != nil {
				return err
			}
			err = os.Chmod(temp.Name(), 0777)
			if err != nil {
				return err
			}
			plugin.args[0] = temp.Name()
		}
	}
	return ExecPlugin(plugin)
}

// ExecPlugin runs a plugin found by the above findPlugin function
func ExecPlugin(info *pluginInfo) error {
	plugin := &exec.Cmd{
		Path:   info.args[0],
		Args:   info.args,
		Stdout: os.Stdout,
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
	}
	return plugin.Run()
}
