package ps1

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/template-color"
)

var (
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
			// (like :wheel_of_dharma: is the closest unicode emoji to the Kubernetes logo)
			//
			// This is the hex form of \u262a so that it can render on older versions of bash/zsh as well
			// https://unicode.org/emoji/charts/full-emoji-list.html#262a
			// https://www.fileformat.info/info/unicode/char/262a/index.htm
			//
			// Inspiration: https://github.com/jonmosco/kube-ps1
			// How to: https://stackoverflow.com/a/37447234/337735
			return "\xe2\x98\xaa ", nil
		},
	}

	// For documentation of supported tokens, see internal/cmd/prompt/command.go
	formatData = func(config *config.Config) (interface{}, error) {
		kcc, err := config.KafkaClusterConfig()
		if err != nil {
			return nil, err
		}
		kafkaClusterId := "(none)"
		kafkaClusterName := "(none)"
		kafkaAPIKey := "(none)"
		if kcc != nil {
			if kcc.ID != "" {
				kafkaClusterId = kcc.ID
			}
			if kcc.Name != "" {
				kafkaClusterName = kcc.Name
			}
			if kcc.APIKey != "" {
				kafkaAPIKey = kcc.APIKey
			}
		}
		return map[string]interface{} {
			"CLIName":          config.CLIName,
			"ContextName":      config.CurrentContext,
			"AccountId":        config.Auth.Account.Id,
			"AccountName":      config.Auth.Account.Name,
			"KafkaClusterId":   kafkaClusterId,
			"KafkaClusterName": kafkaClusterName,
			"KafkaAPIKey":      kafkaAPIKey,
			"UserName":         config.Auth.User.Email,
		}, nil
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
	prompt, err := p.ParseTemplate(result)
	if err != nil {
		return "", err
	}
	return prompt, nil
}

func (p *Prompt) GetFuncs() template.FuncMap {
	m := template_color.GetColorFuncs()
	return m
}

func (p *Prompt) ParseTemplate(text string) (string, error) {
	t, err := template.New("tmpl").Funcs(p.GetFuncs()).Parse(text)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	data, err := formatData(p.Config)
	if err != nil {
		return "", err
	}
	if err := t.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
