package utils

import (
	fColor "github.com/fatih/color"

	"github.com/confluentinc/cli/v3/pkg/color"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func OutputErr(s string) {
	c := fColor.New(color.ErrorColor)
	output.Println(false, c.Sprintf(s))
}

func OutputErrf(s string, args ...any) {
	c := fColor.New(color.ErrorColor)
	output.Printf(false, c.Sprint(s), args...)
}

func OutputInfo(s string) {
	output.Println(false, s)
}

func OutputInfof(s string, args ...any) {
	output.Printf(false, s, args...)
}

func OutputWarn(s string) {
	c := fColor.New(color.WarnColor)
	output.Println(false, c.Sprint(s))
}
