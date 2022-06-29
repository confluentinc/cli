package plugin

import (
	"github.com/confluentinc/cli/internal/pkg/utils"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

func SearchPath() (map[string][]string, error) {
	pluginMap := make(map[string][]string)
	re := regexp.MustCompile(`^confluent(-[a-z][0-9_a-z]*)+(\.[a-z]+)?$`)
	var pathSlice []string
	if runtime.GOOS == "windows" {
		pathSlice = strings.Split(os.Getenv("PATH"), ";")
	} else {
		pathSlice = strings.Split(os.Getenv("PATH"), ":")
	}

	for _, dir := range pathSlice {
		if err := filepath.Walk(dir, pluginWalkFn(re, pluginMap)); err != nil {
			return nil, err
		}
	}
	return pluginMap, nil
}

func pluginWalkFn(re *regexp.Regexp, pluginMap map[string][]string) func(string, fs.FileInfo, error) error {
	return func(path string, info fs.FileInfo, _ error) error {
		pluginName := filepath.Base(path)
		if re.MatchString(pluginName) && ((runtime.GOOS != "windows" && isExecutable(info)) || (runtime.GOOS == "windows" && isExecutableWindows(pluginName))) {
			if strings.Contains(pluginName, ".") {
				pluginName = pluginName[:strings.LastIndex(pluginName, ".")]
			}
			pluginMap[pluginName] = append(pluginMap[pluginName], path)
		}
		return nil
	}
}

func isExecutable(info fs.FileInfo) bool {
	m := info.Mode()
	return !m.IsDir() && m&0111 != 0
}

func isExecutableWindows(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return utils.Contains([]string{".bat", ".cmd", ".com", ".exe", ".ps1"}, ext)
}
