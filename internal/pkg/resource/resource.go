package resource

import (
	"fmt"
	"strings"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2/flink/v2"
	v2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"

	"github.com/confluentinc/cli/internal/pkg/errors"
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

type specResourcePtr interface {
	*cmkv2.CmkV2Cluster | *flinkv2.FcpmV2ComputePool
}

type v2ResourcePtr interface {
	*orgv2.OrgV2Environment | *orgv2.OrgV2Organization | *v2.IamV2IdentityProvider
	GetDisplayName() string
	GetId() string
}

type v2Resource interface {
	orgv2.OrgV2Environment | orgv2.OrgV2Organization | v2.IamV2IdentityProvider
}

func ConvertToPtrSlice[V v2Resource](resources []V) []*V {
	ptrs := make([]*V, len(resources))
	for i := range resources {
		ptrs[i] = &resources[i]
	}
	return ptrs
}

// ConvertV2NameToId ConvertV2NamesToID returns a resource name's corresponding ID or returns the input string if not found
func ConvertV2NameToId[V v2ResourcePtr](input string, resources []V) (string, error) {
	namesToIDs, err := GetV2NamesToIds(resources)
	if err != nil {
		return input, err
	}
	if resourceId, ok := namesToIDs[input]; ok {
		return resourceId, nil
	} else {
		return input, nil
	}
}

// GetV2NamesToIds return a mapping from resource names to their respective IDs
func GetV2NamesToIds[V v2ResourcePtr](resources []V) (map[string]string, error) {
	namesToIDs := make(map[string]string, len(resources))
	for _, resource := range resources {
		if _, ok := namesToIDs[resource.GetDisplayName()]; !ok {
			namesToIDs[resource.GetDisplayName()] = resource.GetId()
		} else {
			return nil, errors.NewErrorWithSuggestions(fmt.Sprintf(`the resource name "%s" is shared across multiple resources`, resource.GetDisplayName()), "instead use resource id")
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
