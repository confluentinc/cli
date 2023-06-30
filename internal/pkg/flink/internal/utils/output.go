package utils

import (
	fColor "github.com/fatih/color"

	"github.com/confluentinc/cli/internal/pkg/color"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func OutputErr(s string) {
	c := fColor.New(color.ErrorColor)
	output.Println(c.Sprintf(s))
}

func OutputErrf(s string, args ...any) {
	c := fColor.New(color.ErrorColor)
	output.Printf(c.Sprint(s), args...)
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
