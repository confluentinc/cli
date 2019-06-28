package ps1

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
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
			if config.Auth == nil || config.Auth.Account == nil || config.Auth.Account.Id == "" {
				return "(none)", nil
			} else {
				return config.Auth.Account.Id, nil
			}
		},
		"%E": func(config *config.Config) (string, error) {
			if config.Auth == nil || config.Auth.Account == nil || config.Auth.Account.Name == "" {
				return "(none)", nil
			} else {
				return config.Auth.Account.Name, nil
			}
		},
		"%k": func(config *config.Config) (string, error) {
			kcc, err := config.KafkaClusterConfig()
			if err != nil && err != errors.ErrNoContext {
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
			if err != nil && err != errors.ErrNoContext {
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
			if err != nil && err != errors.ErrNoContext {
				return "", err
			}
			if kcc == nil || kcc.APIKey == "" {
				return "(none)", nil
			} else {
				return kcc.APIKey, nil
			}
		},
		"%u": func(config *config.Config) (string, error) {
			if config.Auth == nil || config.Auth.User == nil || config.Auth.User.Email == "" {
				return "(none)", nil
			} else {
				return config.Auth.User.Email, nil
			}
		},
	}

	// For documentation of supported tokens, see internal/cmd/prompt/command.go
	formatData = func(config *config.Config) (interface{}, error) {
		kcc, err := config.KafkaClusterConfig()
		if err != nil && err != errors.ErrNoContext {
			return nil, err
		}
		kafkaClusterID := "(none)"
		kafkaClusterName := "(none)"
		kafkaAPIKey := "(none)"
		accountID := "(none)"
		accountName := "(none)"
		userEmail := "(none)"
		if kcc != nil {
			if kcc.ID != "" {
				kafkaClusterID = kcc.ID
			}
			if kcc.Name != "" {
				kafkaClusterName = kcc.Name
			}
			if kcc.APIKey != "" {
				kafkaAPIKey = kcc.APIKey
			}
		}
		if config.Auth != nil {
			if config.Auth.Account != nil {
				if config.Auth.Account.Id != "" {
					accountID = config.Auth.Account.Id
				}
				if config.Auth.Account.Name != "" {
					accountName = config.Auth.Account.Name
				}
			}
			if config.Auth.User != nil {
				if config.Auth.User.Email != "" {
					userEmail = config.Auth.User.Email
				}
			}
		}
		return map[string]interface{}{
			"CLIName":          config.CLIName,
			"ContextName":      config.CurrentContext,
			"AccountId":        accountID,
			"AccountName":      accountName,
			"KafkaClusterId":   kafkaClusterID,
			"KafkaClusterName": kafkaClusterName,
			"KafkaAPIKey":      kafkaAPIKey,
			"UserName":         userEmail,
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
	m["ToUpper"] = strings.ToUpper
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
