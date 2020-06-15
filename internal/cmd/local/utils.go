package local

import (
	"fmt"
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
