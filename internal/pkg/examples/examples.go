package examples

import (
	"fmt"
	"strings"
)

type Example struct {
	Desc string
	Code string
}

func BuildExampleString(examples ...Example) string {
	str := strings.Builder{}
	for _, e := range examples {
		str.WriteString(e.Desc + "\n\n")
		str.WriteString("::\n\n")
		str.WriteString(tab(e.Code) + "\n\n")
	}
	return str.String()
}

func tab(block string) string {
	str := strings.Builder{}
	for _, line := range strings.Split(block, "\n") {
		str.WriteString(fmt.Sprintf("  %s\n", line))
	}
	return strings.TrimSuffix(str.String(), "\n")
}
