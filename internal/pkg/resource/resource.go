package resource

import (
	"strings"
)

type Type int

const (
	Unknown Type = iota
	Cloud
	Kafka
	Ksql
	SchemaRegistry
	ServiceAccount
	User
)

func LookupType(resourceId string) Type {
	if resourceId == "cloud" {
		return Cloud
	}

	prefixToType := map[string]Type{
		"lkc":    Kafka,
		"lksqlc": Ksql,
		"lsrc":   SchemaRegistry,
		"sa":     ServiceAccount,
		"u":      User,
	}

	for prefix, resourceType := range prefixToType {
		if strings.HasPrefix(resourceId, prefix+"-") {
			return resourceType
		}
	}

	return Unknown
}
