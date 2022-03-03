package resource

import (
	"fmt"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	CloudType = "cloud"
	KafkaType = "kafka"
	KsqlType  = "ksql"
	SrType    = "schema-registry"
)

func LookupType(resourceId string) (string, error) {
	if resourceId == CloudType {
		return CloudType, nil
	}

	prefixToType := map[string]string{
		"lkc":    KafkaType,
		"lksqlc": KsqlType,
		"lsrc":   SrType,
	}

	for prefix, resourceType := range prefixToType {
		if strings.HasPrefix(resourceId, prefix+"-") {
			return resourceType, nil
		}
	}

	return "", fmt.Errorf(errors.ResourceNotFoundErrorMsg, resourceId)
}
