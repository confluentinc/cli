package utils

import "strings"

func SplitArgs(s string) []string {
	quotes := ' '
	escaped := false
	split := strings.FieldsFunc(s, func(r rune) bool {

		if !escaped && (r == '\'' || r == '"') {
			if quotes == ' ' {
				quotes = r
			} else if quotes == r {
				quotes = ' '
			}
		}
		if r == '\\' && !escaped {
			escaped = true
		} else {
			escaped = false
		}
		return quotes == ' ' && r == ' '
	})
	for i, arg := range split {
		if len(arg) > 2 &&
			((arg[0] == '"' && arg[len(arg)-1] == '"') ||
				(arg[0] == '\'' && arg[len(arg)-1] == '\'')){
			split[i] = arg[1 : len(arg)-1]
		}
	}
	return split
}
