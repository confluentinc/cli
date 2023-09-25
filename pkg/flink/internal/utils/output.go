package utils

import (
	"strings"

	fColor "github.com/fatih/color"

	"github.com/confluentinc/cli/v3/pkg/color"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func OutputErr(s string) {
	c := fColor.New(color.ErrorColor)
	output.Println(c.Sprintf(s))
}

func OutputErrf(s string, args ...any) {
	c := fColor.New(color.ErrorColor)
	output.Printf(c.Sprintln(s), args...)
}

func OutputInfo(s string) {
	output.Println(s)
}

func OutputInfof(s string, args ...any) {
	output.Printf(s, args...)
}

func OutputWarn(s string) {
	c := fColor.New(color.WarnColor)
	output.Println(c.Sprint(s))
}

func OutputWarnf(s string, args ...any) {
	c := fColor.New(color.WarnColor)
	output.Printf(c.Sprintln(s), args...)
}

func GetMaxStrWidth(str string) int {
	// split the string by lines and find the longest line
	lines := strings.Split(strings.ReplaceAll(str, "\r\n", "\n"), "\n")
	maxWidth := 0
	for idx, line := range lines {
		lineLength := len(line) + 1
		// the last line does not have a extra new line char that needs to be counted
		if idx == len(lines)-1 {
			lineLength = len(line)
		}
		maxWidth = max(maxWidth, lineLength)
	}
	return maxWidth
}
