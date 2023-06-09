package panic

import (
	"strings"
)

type Panic struct {
	ErrorMsg string
}

func (p *Panic) Error() string {
	return p.ErrorMsg
}

// ParseStack formats the stack trace resulting from a panic to only include filenames and line numbers up until panic
func ParseStack(stack string) *[]string {
	trace := strings.Split(stack, "\n")
	if trace[len(trace)-1] == "" {
		trace = trace[:len(trace)-1]
	}
	panicIndex := 0
	for i := range trace {
		trace[i] = strings.TrimSpace(trace[i])
		if strings.Contains(trace[i], "panic.go") {
			panicIndex = i
		}
	}
	trace = trace[panicIndex:]
	var result []string
	for _, s := range trace {
		if strings.Contains(s, ".go") {
			result = append(result, s)
		}
	}
	return &result
}
