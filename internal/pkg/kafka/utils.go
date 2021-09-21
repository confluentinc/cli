package kafka

import (
	"fmt"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

type ListACLsContextKey string

const Requester ListACLsContextKey = "requester"

func ToMap(configs []string) (map[string]string, error) {
	configMap := make(map[string]string)
	for _, cfg := range configs {
		pair := strings.SplitN(cfg, "=", 2)
		if len(pair) < 2 {
			return nil, fmt.Errorf(errors.ConfigurationFormErrorMsg)
		}
		configMap[pair[0]] = pair[1]
	}
	return configMap, nil
}
