package local

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

func BuildTabbedList(arr []string) string {
	var list strings.Builder
	for _, x := range arr {
		fmt.Fprintf(&list, "  %s\n", x)
	}
	return list.String()
}

func ExtractConfig(data []byte) map[string]interface{} {
	re := regexp.MustCompile(`(?m)^[^\s#]*=.+`)
	matches := re.FindAllString(string(data), -1)

	config := make(map[string]interface{})
	for _, match := range matches {
		x := strings.Split(match, "=")
		key, val := x[0], x[1]
		config[key] = val
	}
	return config
}

func CollectFlags(flags *pflag.FlagSet, flagTypes map[string]interface{}) ([]string, error) {
	var args []string

	for key, typeDefault := range flagTypes {
		var val interface{}
		var err error

		switch typeDefault.(type) {
		case bool:
			val, err = flags.GetBool(key)
		case int:
			val, err = flags.GetInt(key)
		case string:
			val, err = flags.GetString(key)
		case []string:
			val, err = flags.GetStringArray(key)
		}
		if err != nil {
			return []string{}, err
		}

		isDefault := false
		switch typeDefault.(type) {
		case bool, int, string:
			isDefault = val == typeDefault
		default:
			isDefault = val == nil
		}
		if isDefault {
			continue
		}

		flag := fmt.Sprintf("--%s", key)

		switch typeDefault.(type) {
		case bool:
			args = append(args, flag)
		case int:
			args = append(args, flag, strconv.Itoa(val.(int)))
		case string:
			args = append(args, flag, val.(string))
		case []string:
			for _, v := range val.([]string) {
				args = append(args, flag, v)
			}
		}
	}

	return args, nil
}

func exists(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}
