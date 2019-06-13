package ps1

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/fatih/color"

	"github.com/confluentinc/cli/internal/pkg/config"
)

var (
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()

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

	// For documentation of supported tokens, see internal/cmd/prompt/command.go
	formatTokens = map[string]func(config *config.Config) (string, error){
		"%c": func(config *config.Config) (string, error) {
			context := config.CurrentContext
			if context == "" {
				context = "(none)"
			}
			return context, nil
		},
		"%e": func(config *config.Config) (string, error) {
			return config.Auth.Account.Id, nil
		},
		"%E": func(config *config.Config) (string, error) {
			return config.Auth.Account.Name, nil
		},
		"%k": func(config *config.Config) (string, error) {
			kcc, err := config.KafkaClusterConfig()
			if err != nil {
				return "", err
			}
			if kcc == nil {
				return "(none)", nil
			} else {
				return kcc.ID, nil
			}
		},
		"%K": func(config *config.Config) (string, error) {
			kcc, err := config.KafkaClusterConfig()
			if err != nil {
				return "", err
			}
			if kcc == nil || kcc.Name == "" {
				return "(none)", nil
			} else {
				return kcc.Name, nil
			}
		},
		"%a": func(config *config.Config) (string, error) {
			kcc, err := config.KafkaClusterConfig()
			if err != nil {
				return "", err
			}
			if kcc == nil || kcc.APIKey == "" {
				return "(none)", nil
			} else {
				return kcc.APIKey, nil
			}
		},
		"%u": func(config *config.Config) (string, error) {
			return config.Auth.User.Email, nil
		},
		"%X": func(config *config.Config) (string, error) {
			// :star_and_crescent: is the closest unicode emoji to the Confluent logo
			// like :wheel_of_dharma: is the closest unicode emoji to the Kubernetes logo
			//
			// This is the hex form of \u262a so that it can render on older versions of bash/zsh as well
			// https://unicode.org/emoji/charts/full-emoji-list.html#262a
			// https://www.fileformat.info/info/unicode/char/262a/index.htm
			//
			// Inspiration: https://github.com/jonmosco/kube-ps1
			return "\xe2\x98\xaa", nil
		},
	}

	colorFunc = func(attr color.Attribute, text ...interface{}) string {
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

	colorLookupFunc = func(name string, m map[string]color.Attribute, key string, text ...interface{}) string {
		v, found := m[key]
		if !found {
			return fmt.Sprintf("#error{%s not found: %s}", name, key)
		}
		return colorFunc(v, text...)
	}

	ColorFuncs = template.FuncMap{
		// inline color funcs
		"fgcolor": func(c string, text ...interface{}) string {
			return colorLookupFunc("fgcolor", fgColors, c, text...)
		},
		"bgcolor": func(c string, text ...interface{}) string {
			return colorLookupFunc("bgcolor", bgColors, c, text...)
		},
		"colorattr": func(a string, text ...interface{}) string {
			return colorLookupFunc("colorattr", colorAttrs, a, text...)
		},
		// color is an alias for fgcolor
		"color": func(c string, text ...interface{}) string {
			return colorLookupFunc("fgcolor", fgColors, c, text...)
		},
		// resetcolor ends all the color attributes
		"resetcolor": func() string {
			return colorFunc(color.Reset)
		},
	}
)

// Prompt outputs context about the current CLI config suitable for a PS1 prompt.
// It allows user configuration by parsing format flags.
type Prompt struct {
	Config *config.Config
}

// Get parses the format string and returns the string with all supported tokens replaced with actual values
func (p *Prompt) Get(format string) (string, error) {
	result := format
	for token, f := range formatTokens {
		v, err := f(p.Config)
		if err != nil {
			return "", err
		}
		result = strings.ReplaceAll(result, token, v)
	}
	envColor, err := p.InferEnvironmentColor()
	if err != nil {
		return "", err
	}
	return envColor(result), nil
}

// inferEnvironmentColor provides a heuristic for determining the environment
// based on the context name, customer environment name, or kafka cluster name.
// (prod=red, stag=yellow, dev=green, unknown=none)
func (p *Prompt) InferEnvironmentColor() (func(a ...interface{}) string, error) {
	noColor := func(a ...interface{}) string {
		return fmt.Sprint(a...)
	}

	envColor := inferColorBasedOnEnvName(p.Config.CurrentContext)
	if envColor != nil {
		return envColor, nil
	}

	envColor = inferColorBasedOnEnvName(p.Config.Auth.Account.Name)
	if envColor != nil {
		return envColor, nil
	}

	kcc, err := p.Config.KafkaClusterConfig()
	if err != nil {
		return nil, err
	}
	if kcc == nil {
		return noColor, nil
	}
	envColor = inferColorBasedOnEnvName(kcc.Name)
	if envColor != nil {
		return envColor, nil
	}

	return noColor, nil
}

func inferColorBasedOnEnvName(name string) func(a ...interface{}) string {
	name = strings.ToLower(name)
	if strings.Contains(name, "prod") || strings.Contains(name, "prd") {
		return red
	}
	if strings.Contains(name, "stag") || strings.Contains(name, "stg") {
		return yellow
	}
	if strings.Contains(name, "dev") || strings.Contains(name, "dev") {
		return green
	}
	return nil
}
