package local

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
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
