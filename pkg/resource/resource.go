package resource

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/types"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

const (
	Unknown                         = "unknown"
	AccessPoint                     = "access point"
	ACL                             = "ACL"
	ApiKey                          = "API key"
	Broker                          = "broker"
	ByokKey                         = "self-managed key"
	ClientQuota                     = "client quota"
	Cloud                           = "cloud"
	ClusterLink                     = "cluster link"
	Connector                       = "connector"
	CustomConnectorPlugin           = "custom connector plugin"
	ConsumerShare                   = "consumer share"
	Context                         = "context"
	Dek                             = "DEK"
	DnsForwarder                    = "DNS forwarder"
	DnsRecord                       = "DNS record"
	Environment                     = "environment"
	Flink                           = "flink"
	FlinkComputePool                = "Flink compute pool"
	FlinkRegion                     = "Flink region"
	FlinkStatement                  = "Flink SQL statement"
	IdentityPool                    = "identity pool"
	IdentityProvider                = "identity provider"
	IpGroup                         = "IP group"
	IpFilter                        = "IP filter"
	KafkaCluster                    = "Kafka cluster"
	Kek                             = "KEK"
	KsqlCluster                     = "KSQL cluster"
	MirrorTopic                     = "mirror topic"
	Network                         = "network"
	NetworkLinkEndpoint             = "network link endpoint"
	NetworkLinkService              = "network link service"
	NetworkLinkServiceAssociation   = "network link service association"
	Organization                    = "organization"
	Peering                         = "peering"
	PrivateLinkAccess               = "private link access"
	PrivateLinkAttachment           = "private link attachment"
	PrivateLinkAttachmentConnection = "private link attachment connection"
	ProviderShare                   = "provider share"
	Pipeline                        = "pipeline"
	SchemaExporter                  = "schema exporter"
	SchemaRegistryCluster           = "Schema Registry cluster"
	SchemaRegistryConfiguration     = "Schema Registry configuration"
	ServiceAccount                  = "service account"
	SsoGroupMapping                 = "SSO group mapping"
	Topic                           = "topic"
	TransitGatewayAttachment        = "transit gateway attachment"
	User                            = "user"
)

const (
	AccessPointPrefix           = "ap"
	ConnectorPrefix             = "lcc"
	DnsRecordPrefix             = "dnsrec"
	EnvironmentPrefix           = "env"
	IdentityPoolPrefix          = "pool"
	IdentityProviderPrefix      = "op"
	FlinkComputePoolPrefix      = "lfcp"
	KafkaClusterPrefix          = "lkc"
	KsqlClusterPrefix           = "lksqlc"
	SchemaRegistryClusterPrefix = "lsrc"
	ServiceAccountPrefix        = "sa"
	SsoGroupMappingPrefix       = "group"
	UserPrefix                  = "u"
)

var prefixToResource = map[string]string{
	AccessPointPrefix:           AccessPoint,
	ConnectorPrefix:             Connector,
	DnsRecordPrefix:             DnsRecord,
	EnvironmentPrefix:           Environment,
	IdentityPoolPrefix:          IdentityPool,
	IdentityProviderPrefix:      IdentityProvider,
	FlinkComputePoolPrefix:      FlinkComputePool,
	KafkaClusterPrefix:          KafkaCluster,
	KsqlClusterPrefix:           KsqlCluster,
	SchemaRegistryClusterPrefix: SchemaRegistryCluster,
	ServiceAccountPrefix:        ServiceAccount,
	SsoGroupMappingPrefix:       SsoGroupMapping,
	UserPrefix:                  User,
}

var resourceToPrefix = map[string]string{
	AccessPoint:           AccessPointPrefix,
	DnsRecord:             DnsRecordPrefix,
	Environment:           EnvironmentPrefix,
	IdentityPool:          IdentityPoolPrefix,
	IdentityProvider:      IdentityProviderPrefix,
	KafkaCluster:          KafkaClusterPrefix,
	KsqlCluster:           KsqlClusterPrefix,
	SchemaRegistryCluster: SchemaRegistryClusterPrefix,
	ServiceAccount:        ServiceAccountPrefix,
	SsoGroupMapping:       SsoGroupMappingPrefix,
	User:                  UserPrefix,
}

func LookupType(id string) string {
	if id == Cloud || id == Flink {
		return id
	}

	if x := strings.SplitN(id, "-", 2); len(x) == 2 {
		prefix := x[0]
		if resource, ok := prefixToResource[prefix]; ok {
			return resource
		}
	}

	return Unknown
}

func ValidatePrefixes(resourceType string, args []string) error {
	prefix, ok := resourceToPrefix[resourceType]
	if !ok {
		return nil
	}

	// old group mappings may still have "pool-" instead of "group-"
	// so we must skip the check for this resource
	if prefix == SsoGroupMappingPrefix {
		return nil
	}

	var malformed []string
	for _, id := range args {
		if LookupType(id) != resourceType {
			malformed = append(malformed, id)
		}
	}

	if len(malformed) == 1 {
		return fmt.Errorf(`failed parsing resource ID %s: missing prefix "%s-"`, malformed[0], prefix)
	} else if len(malformed) > 1 {
		return fmt.Errorf(`failed parsing resource IDs %s: missing prefix "%s-"`, utils.ArrayToCommaDelimitedString(malformed, "and"), prefix)
	}

	return nil
}

func ValidateArgs(c *cobra.Command, args []string, resourceType string, checkExistence func(string) bool) error {
	var invalidArgs []string
	for _, arg := range args {
		if !checkExistence(arg) {
			invalidArgs = append(invalidArgs, arg)
		}
	}

	if len(invalidArgs) != 0 {
		return ResourcesNotFoundError(c, resourceType, invalidArgs...)
	}

	return nil
}

func ResourcesNotFoundError(cmd *cobra.Command, resourceType string, invalidArgs ...string) error {
	notFoundErrorMsg := `%s %s not found`
	invalidArgsErrMsg := fmt.Sprintf(notFoundErrorMsg, resourceType, utils.ArrayToCommaDelimitedString(invalidArgs, "and"))
	if len(invalidArgs) > 1 {
		invalidArgsErrMsg = fmt.Sprintf(notFoundErrorMsg, Plural(resourceType), utils.ArrayToCommaDelimitedString(invalidArgs, "and"))
	}

	// Find the full parent command string for use in the suggestion message
	var fullParentCommand string
	if cmd.HasParent() {
		fullParentCommand = cmd.Parent().Name()
		cmd = cmd.Parent()
	}
	for cmd.HasParent() {
		fullParentCommand = fmt.Sprintf("%s %s", cmd.Parent().Name(), fullParentCommand)
		cmd = cmd.Parent()
	}
	invalidResourceSuggestion := fmt.Sprintf(errors.ListResourceSuggestions, resourceType, fullParentCommand)

	return errors.NewErrorWithSuggestions(invalidArgsErrMsg, invalidResourceSuggestion)
}

func Plural(resource string) string {
	if resource == "" {
		return ""
	}

	// Singular words ending w/ these suffixes generally add an extra -es syllable in their plural forms
	var pluralExtraSyllableSuffix = types.NewSet("s", "x", "z", "ch", "sh")

	for suffix := range pluralExtraSyllableSuffix {
		if strings.HasSuffix(resource, suffix) {
			return resource + "es"
		}
	}

	return resource + "s"
}
