package plugin

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/types"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

var pluginRegex = regexp.MustCompile(`^confluent(-[a-z][0-9_a-z]*)+$`)

type pluginInfo struct {
	args     []string
	name     string
	nameSize int
}

// SearchPath goes through the files in the user's $PATH and checks if they are plugins
func SearchPath(cfg *v1.Config) map[string][]string {
	log.CliLogger.Debugf("Searching `$PATH` and `~/.confluent/plugins` for plugins. Plugins can be disabled in %s.\n", cfg.GetFilename())

	pathDirList := filepath.SplitList(os.Getenv("PATH"))
	pluginDir := filepath.Join(os.Getenv("HOME"), ".confluent", "plugins")
	if !types.Contains(pathDirList, pluginDir) {
		pathDirList = append(pathDirList, pluginDir)
	}

	plugins := make(map[string][]string)
	for _, dir := range pathDirList {
		entries, err := os.ReadDir(dir)
		if err != nil {
			log.CliLogger.Warnf("unable to read directory from `$PATH`: %s", dir)
			continue
		}

		for _, entry := range entries {
			if name := PluginFromEntry(entry); name != "" {
				path := filepath.Join(dir, entry.Name())
				plugins[name] = append(plugins[name], path)
			}
		}
	}

	return plugins
}

func PluginFromEntry(entry os.DirEntry) string {
	if !isExecutable(entry) {
		return ""
	}

	name := entry.Name()
	name = strings.TrimSuffix(name, filepath.Ext(name))

	if !pluginRegex.MatchString(name) {
		return ""
	}

	return name
}

func isExecutable(entry fs.DirEntry) bool {
	if runtime.GOOS == "windows" {
		extension := strings.ToUpper(filepath.Ext(entry.Name()))
		return types.Contains(filepath.SplitList(os.Getenv("PATHEXT")), extension)
	}

	fileInfo, err := entry.Info()
	if err != nil {
		return false
	}

	return !fileInfo.Mode().IsDir() && fileInfo.Mode()&0111 != 0
}

// FindPlugin determines if the arguments passed in are meant for a plugin
func FindPlugin(cmd *cobra.Command, args []string, cfg *v1.Config) *pluginInfo {
	pluginMap := SearchPath(cfg)

	plugin := newPluginInfo(args)
	for len(plugin.name) > len(pversion.CLIName) {
		if pluginPathList, ok := pluginMap[plugin.name]; ok {
			if cmd, _, _ := cmd.Find(args); strings.ReplaceAll(cmd.CommandPath(), " ", "-") == plugin.name {
				log.CliLogger.Warnf("[WARN] User plugin %s is ignored because its command line invocation matches existing CLI command `%s`.\n", pluginPathList[0], cmd.CommandPath())
				break
			}
			plugin.args = append([]string{pluginPathList[0]}, plugin.args...)
			return plugin
		}
		plugin.args = append([]string{args[plugin.nameSize-1]}, plugin.args...)
		plugin.nameSize--
		plugin.name = plugin.name[:strings.LastIndex(plugin.name, "-")]
	}
	return nil
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
		args:     infoArgs,
		name:     strings.Join(name, "-"),
		nameSize: len(name) - 1,
	}
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
