package local

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
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

func versionCmp(a, b string) (int, error) {
	as, err := stringSliceToIntSlice(strings.Split(a, "."))
	if err != nil {
		return 0, err
	}
	bs, err := stringSliceToIntSlice(strings.Split(b, "."))
	if err != nil {
		return 0, err
	}

	for i := len(as); i < len(bs); i++ {
		as = append(as, 0)
	}
	for i := len(bs); i < len(as); i++ {
		bs = append(bs, 0)
	}

	for i := 0; i < len(as) && i < len(bs); i++ {
		if as[i] == bs[i] {
			continue
		}
		return as[i] - bs[i], nil
	}

	return 0, nil
}

func stringSliceToIntSlice(arr []string) ([]int, error) {
	out := make([]int, len(arr))
	for i, a := range arr {
		n, err := strconv.Atoi(a)
		if err != nil {
			return []int{}, err
		}
		out[i] = n
	}
	return out, nil
}
