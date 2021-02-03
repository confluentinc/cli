package utils

import "strings"

func SplitArgs(s string) []string {
	quotation := false
	splitted := strings.FieldsFunc(s, func(r rune) bool {
		if r == '"' {
			quotation = !quotation
		}
		return !quotation && r == ' '
	})
	for i, arg := range splitted {
		if len(arg) > 2 && arg[0] == '"' && arg[len(arg)-1] == '"' {
			splitted[i] = arg[1 : len(arg)-1]
		}
	}
	return splitted
}
