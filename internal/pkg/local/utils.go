package local

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

func BuildTabbedList(arr []string) string {
	sort.Strings(arr)

	var list strings.Builder
	for _, x := range arr {
		fmt.Fprintf(&list, "  %s\n", x)
	}
	return list.String()
}

func ExtractConfig(data []byte) map[string]string {
	re := regexp.MustCompile(`(?m)^[^\s#]*=.+`)
	matches := re.FindAllString(string(data), -1)
	config := map[string]string{}

	for _, match := range matches {
		x := strings.Split(match, "=")
		key, val := x[0], x[1]
		config[key] = val
	}
	return config
}

func Contains(haystack []string, needle string) bool {
	for _, x := range haystack {
		if x == needle {
			return true
		}
	}
	return false
}

func CollectFlags(flags *pflag.FlagSet, flagTypes map[string]interface{}) ([]string, error) {
	var args []string

	for key, typeDefault := range flagTypes {
		var val interface{}
		var err error

		switch typeDefault.(type) {
		case bool:
			val, err = flags.GetBool(key)
		case string:
			val, err = flags.GetString(key)
		case int:
			val, err = flags.GetInt(key)
		}

		if err != nil {
			return []string{}, err
		}
		if val == typeDefault {
			continue
		}

		flag := fmt.Sprintf("--%s", key)

		switch typeDefault.(type) {
		case bool:
			args = append(args, flag)
		case string:
			args = append(args, flag, val.(string))
		case int:
			args = append(args, flag, strconv.Itoa(val.(int)))
		}
	}

	return args, nil
}

func exists(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}
