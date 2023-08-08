package resource

import (
	"strings"
)

const (
	Unknown               = "unknown"
	ApiKey                = "API key"
	ByokKey               = "self-managed key"
	ClientQuota           = "client quota"
	Cloud                 = "cloud"
	ClusterLink           = "cluster link"
	Connector             = "connector"
	ConsumerShare         = "consumer share"
	Context               = "context"
	Environment           = "environment"
	FlinkComputePool      = "Flink compute pool"
	FlinkRegion           = "Flink region"
	FlinkIamBinding       = "Flink IAM binding"
	FlinkStatement        = "Flink SQL statement"
	IdentityPool          = "identity pool"
	IdentityProvider      = "identity provider"
	KafkaCluster          = "Kafka cluster"
	KsqlCluster           = "KSQL cluster"
	MirrorTopic           = "mirror topic"
	Organization          = "organization"
	ProviderShare         = "provider share"
	Pipeline              = "pipeline"
	SchemaExporter        = "schema exporter"
	SchemaRegistryCluster = "Schema Registry cluster"
	ServiceAccount        = "service account"
	SsoGroupMapping       = "SSO group mapping"
	Topic                 = "topic"
	User                  = "user"
)

const (
	ClusterLinkPrefix           = "link"
	EnvironmentPrefix           = "env"
	IdentityPoolPrefix          = "pool"
	IdentityProviderPrefix      = "op"
	KafkaClusterPrefix          = "lkc"
	KsqlClusterPrefix           = "lksqlc"
	SchemaRegistryClusterPrefix = "lsrc"
	ServiceAccountPrefix        = "sa"
	UserPrefix                  = "u"
)

var prefixToResource = map[string]string{
	ClusterLinkPrefix:           ClusterLink,
	EnvironmentPrefix:           Environment,
	IdentityPoolPrefix:          IdentityPool,
	IdentityProviderPrefix:      IdentityProvider,
	KafkaClusterPrefix:          KafkaCluster,
	KsqlClusterPrefix:           KsqlCluster,
	SchemaRegistryClusterPrefix: SchemaRegistryCluster,
	ServiceAccountPrefix:        ServiceAccount,
	UserPrefix:                  User,
}

func LookupType(resourceId string) string {
	if resourceId == Cloud {
		return Cloud
	}

	if x := strings.SplitN(resourceId, "-", 2); len(x) == 2 {
		prefix := x[0]
		if resource, ok := prefixToResource[prefix]; ok {
			return resource
		}
	}

	return Unknown
}
