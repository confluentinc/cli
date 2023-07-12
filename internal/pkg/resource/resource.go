package resource

import (
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2/flink/v2"
	v2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
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

type specResource interface {
	cmkv2.CmkV2Cluster | flinkv2.FcpmV2ComputePool
}

type v2Resource interface {
	orgv2.OrgV2Environment | orgv2.OrgV2Organization | v2.IamV2IdentityProvider
	GetDisplayName() string
	GetID() string
}

func ConvertNameToId[V v2Resource](resourceId string, resources []V) (string, error) {
	return "", nil
}

// GetNamesToIDs return a mapping from resource names to their respective IDs
func GetNamesToIDs[V v2Resource](resources []V) (map[string]string, error) {
	namesToIDs := make(map[string]string, len(resources))
	for _, resource := range resources {
		if _, ok := namesToIDs[resource.GetDisplayName()]; !ok {
			namesToIDs[resource.GetDisplayName()] = resource.GetID()
		} else {
			return nil, errors.New("Duplicate name")
		}
	}
	return namesToIDs, nil
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
