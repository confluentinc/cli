package template_color

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/fatih/color"
)

var (
	fgColors = map[string]color.Attribute{
		"black":   color.FgBlack,
		"red":     color.FgRed,
		"green":   color.FgGreen,
		"yellow":  color.FgYellow,
		"blue":    color.FgBlue,
		"magenta": color.FgMagenta,
		"cyan":    color.FgCyan,
		"white":   color.FgWhite,
	}

	bgColors = map[string]color.Attribute{
		"black":   color.BgBlack,
		"red":     color.BgRed,
		"green":   color.BgGreen,
		"yellow":  color.BgYellow,
		"blue":    color.BgBlue,
		"magenta": color.BgMagenta,
		"cyan":    color.BgCyan,
		"white":   color.BgWhite,
	}

	colorAttrs = map[string]color.Attribute{
		"bold":      color.Bold,
		"underline": color.Underline,
		"invert":    color.ReverseVideo,
	}

	funcMap = template.FuncMap{
		// color is an alias for fgcolor
		"color": func(c string, text ...interface{}) string {
			return FgColor(c, text...)
		},
		"fgcolor": func(c string, text ...interface{}) string {
			return FgColor(c, text...)
		},
		"bgcolor": func(c string, text ...interface{}) string {
			return BgColor(c, text...)
		},
		"colorattr": func(a string, text ...interface{}) string {
			return ColorAttr(a, text...)
		},
		// resetcolor ends all the color attributes
		"resetcolor": func() string {
			return Reset()
		},
	}
)

// GetColorFuncs returns a map of functions used for controlling colors.
//
// * {{color "<color>" "some text"}}
// * {{fgcolor "<color>" "some text"}}
// * {{bgcolor "color>" "some text"}}
// * {{colorattr "<attr>" "some text"}}
//
// Available colors: black, red, green, yellow, blue, magenta, cyan, white
// Available attributes: bold, underline, invert (swaps the fg/bg colors)
//
// Examples:
//
// * {{color "red" "some text" | colorattr "bold" | bgcolor "blue"}}
// * {{color "red"}} some text here {{resetcolor}}
//
// Notes:
//
// * 'color' is just an alias of 'fgcolor'
// * 'resetcolor' will reset all color attributes, not just the most recently set
// * The returned map value is safe to mutate (it's a fresh copy for each call).
func GetColorFuncs() template.FuncMap {
	m := template.FuncMap{}
	for name, impl := range funcMap {
		m[name] = impl
	}
	return m
}

// FgColor wraps text with the foreground color given by name.
// If no text is provided, it just returns the initial format string and you must call Reset() later yourself.
func FgColor(name string, text ...interface{}) string {
	return colorLookupFunc("fgcolor", fgColors, name, text...)
}

// BgColor wraps text with the background color given by name.
// If no text is provided, it just returns the initial format string and you must call Reset() later yourself.
func BgColor(name string, text ...interface{}) string {
	return colorLookupFunc("bgcolor", bgColors, name, text...)
}

// ColorAttr wraps text with the color attribute given by name.
// If no text is provided, it just returns the initial format string and you must call Reset() later yourself.
func ColorAttr(name string, text ...interface{}) string {
	return colorLookupFunc("colorattr", colorAttrs, name, text...)
}

// Reset returns the code to reset all color attributes (fg color, bg color, color attrs, etc)
func Reset() string {
	return colorFunc(color.Reset)
}

func colorFunc(attr color.Attribute, text ...interface{}) string {
	// inline format
	if len(text) > 0 {
		return color.New(attr).Sprint(text...)
	}
	// block format
	buf := &bytes.Buffer{}
	_, _ = color.New(attr).Fprint(buf)
	s := buf.String()
	return s[0 : len(s)-4]
}

func colorLookupFunc(name string, m map[string]color.Attribute, key string, text ...interface{}) string {
	v, found := m[key]
	if !found {
		return fmt.Sprintf("#error{%s not found: %s}", name, key)
	}
	return colorFunc(v, text...)
}
