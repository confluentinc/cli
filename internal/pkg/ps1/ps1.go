package ps1

import (
	"bytes"
	"strings"
	"text/template"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	template_color "github.com/confluentinc/cli/internal/pkg/template-color"
)

var (
	// For documentation of supported tokens, see internal/cmd/prompt/command.go
	formatTokens = map[string]func(cfg *v3.Config) (string, error){
		"%c": func(config *v3.Config) (string, error) {
			context := config.CurrentContext
			if context == "" {
				context = "(none)"
			}
			return context, nil
		},
		"%e": func(config *v3.Config) (string, error) {
			context := config.Context()
			if context == nil {
				return "(none)", nil
			}
			state := context.State
			if state.Auth == nil || state.Auth.Account == nil || state.Auth.Account.Id == "" {
				return "(none)", nil
			} else {
				return state.Auth.Account.Id, nil
			}
		},
		"%E": func(config *v3.Config) (string, error) {
			context := config.Context()
			if context == nil {
				return "(none)", nil
			}
			state := context.State
			if state.Auth == nil || state.Auth.Account == nil || state.Auth.Account.Name == "" {
				return "(none)", nil
			} else {
				return state.Auth.Account.Name, nil
			}
		},
		"%k": func(config *v3.Config) (string, error) {
			context := config.Context()
			if context == nil {
				return "(none)", nil
			}
			kcc := context.KafkaClusterContext.GetActiveKafkaClusterConfig()
			if kcc == nil {
				return "(none)", nil
			} else {
				return kcc.ID, nil
			}
		},
		"%K": func(config *v3.Config) (string, error) {
			context := config.Context()
			if context == nil {
				return "(none)", nil
			}
			kcc := context.KafkaClusterContext.GetActiveKafkaClusterConfig()
			if kcc == nil || kcc.Name == "" {
				return "(none)", nil
			} else {
				return kcc.Name, nil
			}
		},
		"%a": func(config *v3.Config) (string, error) {
			context := config.Context()
			if context == nil {
				return "(none)", nil
			}
			kcc := context.KafkaClusterContext.GetActiveKafkaClusterConfig()
			if kcc == nil || kcc.APIKey == "" {
				return "(none)", nil
			} else {
				return kcc.APIKey, nil
			}
		},
		"%u": func(config *v3.Config) (string, error) {
			context := config.Context()
			if context == nil {
				return "(none)", nil
			}
			state := context.State
			if state.Auth == nil || state.Auth.User == nil || state.Auth.User.Email == "" {
				return "(none)", nil
			} else {
				return state.Auth.User.Email, nil
			}
		},
	}

	// For documentation of supported tokens, see internal/cmd/prompt/command.go
	formatData = func(cfg *v3.Config) (interface{}, error) {
		context := cfg.Context()
		var kcc *v1.KafkaClusterConfig
		if context != nil {
			kcc = context.KafkaClusterContext.GetActiveKafkaClusterConfig()
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
		var state *v2.ContextState
		if context != nil {
			state = context.State
		}
		if state != nil && state.Auth != nil {
			if state.Auth.Account != nil {
				if state.Auth.Account.Id != "" {
					accountID = state.Auth.Account.Id
				}
				if state.Auth.Account.Name != "" {
					accountName = state.Auth.Account.Name
				}
			}
			if state.Auth.User != nil {
				if state.Auth.User.Email != "" {
					userEmail = state.Auth.User.Email
				}
			}
		}
		return map[string]interface{}{
			"CLIName":          cfg.CLIName,
			"ContextName":      cfg.CurrentContext,
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
	Config *v3.Config
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
