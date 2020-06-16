package local

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

func buildTabbedList(slice []string) string {
	sort.Strings(slice)

	var list strings.Builder
	for _, x := range slice {
		fmt.Fprintf(&list, "  %s\n", x)
	}
	return list.String()
}

func extractConfig(data []byte) map[string]string {
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
