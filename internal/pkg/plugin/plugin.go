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
	re := regexp.MustCompile(`confluent(-[a-z]+)+(\.[a-z]+)?`)
	pathSlice := strings.Split(os.Getenv("PATH"), ":")

	for _, dir := range pathSlice {
		err := filepath.Walk(dir, isPluginFn(re, pluginMap))
		if err != nil {
			return nil, err
		}
	}
	return pluginMap, nil
}

func isPluginFn(re *regexp.Regexp, pluginMap map[string][]string) func(string, fs.FileInfo, error) error {
	return func(path string, info fs.FileInfo, _ error) error {
		if re.MatchString(path) && isExec(info) {
			pluginName := filepath.Base(path)
			pluginMap[pluginName] = append(pluginMap[pluginName], path)
		}
		return nil
	}
}

func isExec(info fs.FileInfo) bool {
	if runtime.GOOS == "windows" {
		fileExt := strings.ToLower(filepath.Ext(info.Name()))
		return utils.Contains([]string{".bat", ".cmd", ".com", ".exe", ".ps1"}, fileExt)
	}
	m := info.Mode()
	return !m.IsDir() && m&0111 != 0
}
