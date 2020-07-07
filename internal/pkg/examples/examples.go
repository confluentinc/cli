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
		str.WriteString(e.Desc + "\n")
		str.WriteString(fmt.Sprintf("```%s```\n", e.Code))
	}
	return str.String()
}
