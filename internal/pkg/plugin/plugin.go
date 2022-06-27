package plugin

import (
	"fmt"
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
	re := regexp.MustCompile(`^confluent(-[a-z][0-9a-z]*)+(\.[a-z]+)?$`)
	pathSlice := strings.Split(os.Getenv("PATH"), ":")

	for _, dir := range pathSlice {
		err := filepath.Walk(dir, pluginWalkFn(re, pluginMap))
		if err != nil {
			return nil, err
		}
	}
	return pluginMap, nil
}

func pluginWalkFn(re *regexp.Regexp, pluginMap map[string][]string) func(string, fs.FileInfo, error) error {
	return func(path string, info fs.FileInfo, _ error) error {
		pluginName := filepath.Base(path)
		fmt.Println(path, pluginName, re.MatchString(pluginName))
		if runtime.GOOS == "windows" {
			fmt.Println(isExecutableWindows(pluginName))
		}
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
	fileExt := strings.ToLower(filepath.Ext(name))
	return utils.Contains([]string{".bat", ".cmd", ".com", ".exe", ".ps1"}, fileExt)
}
