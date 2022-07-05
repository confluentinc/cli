package plugin

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/utils"
)

func SearchPath() (map[string][]string, error) {
	pluginMap := make(map[string][]string)
	re := regexp.MustCompile(`^confluent(-[a-z][0-9_a-z]*)+(\.[a-z]+)?$`)
	delimiter := ":"
	if runtime.GOOS == "windows" {
		delimiter = ";"
	}
	pathSlice := strings.Split(os.Getenv("PATH"), delimiter)

	for _, dir := range pathSlice {
		if err := filepath.WalkDir(dir, pluginWalkFn(re, pluginMap)); err != nil {
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
