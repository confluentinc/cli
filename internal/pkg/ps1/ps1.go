package ps1

import (
	"fmt"
	"strings"

	"github.com/fatih/color"

	"github.com/confluentinc/cli/internal/pkg/config"
)

var (
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()

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
			kcc, err := getKafkaClusterConfig(config)
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
			kcc, err := getKafkaClusterConfig(config)
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
			kcc, err := getKafkaClusterConfig(config)
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
	return result, nil
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
	kcc, err := getKafkaClusterConfig(p.Config)
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

func getKafkaClusterConfig(config *config.Config) (*config.KafkaClusterConfig, error) {
	context, err := config.Context()
	if err != nil {
		return nil, err
	}
	kafka := context.Kafka
	if kafka == "" {
		return nil, nil
	} else {
		return context.KafkaClusters[kafka], nil
	}
}
