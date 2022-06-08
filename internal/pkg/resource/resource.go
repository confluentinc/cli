package resource

import (
	"strings"
)

const (
	Unknown          = "unknown"
	Cloud            = "cloud"
	Kafka            = "kafka"
	Ksql             = "ksql"
	SchemaRegistry   = "schema-registry"
	ServiceAccount   = "service-account"
	User             = "user"
	IdentityProvider = "identity-provider"
	IdentityPool     = "identity-pool"

	IdentityPoolPrefix     = "pool"
	IdentityProviderPrefix = "op"
	UserPrefix             = "u"
)

func LookupType(resourceId string) string {
	if resourceId == "cloud" {
		return Cloud
	}

	prefixToType := map[string]string{
		IdentityPoolPrefix:     IdentityPool,
		IdentityProviderPrefix: IdentityProvider,
		"lkc":                  Kafka,
		"lksqlc":               Ksql,
		"lsrc":                 SchemaRegistry,
		"sa":                   ServiceAccount,
		"u":                    User,
	}

	for prefix, resourceType := range prefixToType {
		if strings.HasPrefix(resourceId, prefix+"-") {
			return resourceType
		}
	}

	return Unknown
}
