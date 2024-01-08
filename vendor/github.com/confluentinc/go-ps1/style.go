package ps1

import (
	"fmt"
	"text/template"

	"github.com/fatih/color"
)

var colorFuncs = template.FuncMap{
	"fgcolor": fgColor,
	"bgcolor": bgColor,
	"attr":    attr,
}

func fgColor(name string, text ...interface{}) (string, error) {
	fgColors := map[string]color.Attribute{
		"black":   color.FgBlack,
		"blue":    color.FgBlue,
		"cyan":    color.FgCyan,
		"green":   color.FgGreen,
		"magenta": color.FgMagenta,
		"red":     color.FgRed,
		"white":   color.FgWhite,
		"yellow":  color.FgYellow,
	}

	return printWithAttr("fgcolor", fgColors, name, text...)
}

func bgColor(name string, text ...interface{}) (string, error) {
	bgColors := map[string]color.Attribute{
		"black":   color.BgBlack,
		"blue":    color.BgBlue,
		"cyan":    color.BgCyan,
		"green":   color.BgGreen,
		"magenta": color.BgMagenta,
		"red":     color.BgRed,
		"white":   color.BgWhite,
		"yellow":  color.BgYellow,
	}

	return printWithAttr("bgcolor", bgColors, name, text...)
}

func attr(name string, text ...interface{}) (string, error) {
	attrs := map[string]color.Attribute{
		"bold":      color.Bold,
		"invert":    color.ReverseVideo,
		"italicize": color.Italic,
		"underline": color.Underline,
	}

	return printWithAttr("attr", attrs, name, text...)
}

func printWithAttr(attrType string, attrs map[string]color.Attribute, attr string, text ...interface{}) (string, error) {
	val, ok := attrs[attr]
	if !ok {
		return "", fmt.Errorf(`%s "%s" not found`, attrType, attr)
	}
	return color.New(val).Sprint(text...), nil
}
