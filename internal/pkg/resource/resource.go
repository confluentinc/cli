package resource

import (
	"strings"
)

const (
	Unknown                  = "unknown"
	Acl                      = "ACL"
	ApiKey                   = "API key"
	Broker                   = "broker"
	BrokerConfiguration      = "broker configuration"
	BrokerTask               = "broker task"
	ClientQuota              = "client quota"
	Cloud                    = "cloud"
	ClusterLink              = "cluster link"
	ClusterLinkConfiguration = "cluster link configuration"
	Connector                = "connector"
	Consumer                 = "consumer"
	ConsumerShare            = "consumer share"
	Context                  = "context"
	Environment              = "environment"
	ExporterConfiguration    = "exporter configuration"
	IdentityPool             = "identity pool"
	IdentityProvider         = "identity provider"
	Invitation               = "invitation"
	KafkaCluster             = "Kafka cluster"
	KsqlCluster              = "kSQL cluster"
	MirrorTopic              = "mirror topic"
	Plugin                   = "plugin"
	Pool                     = "pool"
	Price                    = "price"
	PromoCode                = "promo code"
	ProviderShare            = "provider share"
	Region                   = "region"
	Replica                  = "replica"
	RoleBinding              = "role binding"
	SchemaExporter           = "schema exporter"
	SchemaRegistryCluster    = "Schema Registry cluster"
	ServiceAccount           = "service account"
	Topic                    = "topic"
	User                     = "user"
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
	if resourceId == "cloud" {
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
