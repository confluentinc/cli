package resource

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/plural"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

const (
	Unknown                          = "unknown"
	AccessPoint                      = "access point"
	ACL                              = "ACL"
	ApiKey                           = "API key"
	Broker                           = "broker"
	ByokKey                          = "self-managed key"
	CatalogIntegration               = "catalog integration"
	CCPMCustomConnectorPlugin        = "CCPM custom connector plugin"
	CCPMCustomConnectorPluginVersion = "CCPM custom connector plugin version"
	CertificateAuthority             = "certificate authority"
	CertificatePool                  = "certificate pool"
	ClientQuota                      = "client quota"
	Cloud                            = "cloud"
	ClusterLink                      = "cluster link"
	Connector                        = "connector"
	ConnectArtifact                  = "Connect Artifact"
	CustomConnectorPlugin            = "custom connector plugin"
	CustomConnectorPluginVersion     = "custom connector plugin version"
	CustomCodeLogging                = "custom code logging"
	ConsumerShare                    = "consumer share"
	Context                          = "context"
	Dek                              = "DEK"
	DnsForwarder                     = "DNS forwarder"
	DnsRecord                        = "DNS record"
	Environment                      = "environment"
	Flink                            = "flink"
	FlinkArtifact                    = "Flink artifact"
	FlinkConnectivityType            = "Flink connectivity type"
	FlinkComputePool                 = "Flink compute pool"
	FlinkSavepoint                   = "Flink Savepoint"
	FlinkDetachedSavepoint           = "Flink Detached Savepoint"
	FlinkApplication                 = "Flink application"
	FlinkCatalog                     = "Flink catalog"
	FlinkEnvironment                 = "Flink environment"
	FlinkRegion                      = "Flink region"
	FlinkEndpoint                    = "Flink endpoint"
	FlinkStatement                   = "Flink SQL statement"
	FlinkConnection                  = "Flink connection"
	Gateway                          = "gateway"
	IdentityPool                     = "identity pool"
	IdentityProvider                 = "identity provider"
	IpGroup                          = "IP group"
	IpFilter                         = "IP filter"
	KafkaCluster                     = "Kafka cluster"
	Kek                              = "KEK"
	KsqlCluster                      = "KSQL cluster"
	MirrorTopic                      = "mirror topic"
	Network                          = "network"
	NetworkLinkEndpoint              = "network link endpoint"
	NetworkLinkService               = "network link service"
	NetworkLinkServiceAssociation    = "network link service association"
	Organization                     = "organization"
	Peering                          = "peering"
	Plugin                           = "plugin"
	PrivateLinkAccess                = "private link access"
	PrivateLinkAttachment            = "private link attachment"
	PrivateLinkAttachmentConnection  = "private link attachment connection"
	ProviderIntegration              = "provider integration"
	ProviderShare                    = "provider share"
	Pipeline                         = "pipeline"
	SchemaExporter                   = "schema exporter"
	SchemaRegistryCluster            = "Schema Registry cluster"
	SchemaRegistryConfiguration      = "Schema Registry configuration"
	ServiceAccount                   = "service account"
	SsoGroupMapping                  = "SSO group mapping"
	Tableflow                        = "tableflow"
	Topic                            = "topic"
	TransitGatewayAttachment         = "transit gateway attachment"
	User                             = "user"
	UsmKafkaCluster                  = "USM Kafka cluster"
	UsmConnectCluster                = "USM Connect cluster"
)

const (
	AccessPointPrefix           = "ap"
	CatalogIntegrationPrefix    = "tci"
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
	UsmKafkaClusterPrefix       = "usmkc"
	UsmConnectClusterPrefix     = "usmcc"
	CustomCodeLoggingPrefix     = "ccl"
)

var prefixToResource = map[string]string{
	AccessPointPrefix:           AccessPoint,
	CatalogIntegrationPrefix:    CatalogIntegration,
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
	UsmKafkaClusterPrefix:       UsmKafkaCluster,
	UsmConnectClusterPrefix:     UsmConnectCluster,
	CustomCodeLoggingPrefix:     CustomCodeLogging,
}

var resourceToPrefix = map[string]string{
	AccessPoint:           AccessPointPrefix,
	CatalogIntegration:    CatalogIntegrationPrefix,
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
	UsmKafkaCluster:       UsmKafkaClusterPrefix,
	UsmConnectCluster:     UsmConnectClusterPrefix,
	CustomCodeLogging:     CustomCodeLoggingPrefix,
}

func LookupType(id string) string {
	if id == Cloud || id == Flink || id == Tableflow {
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
		invalidArgsErrMsg = fmt.Sprintf(notFoundErrorMsg, plural.Plural(resourceType), utils.ArrayToCommaDelimitedString(invalidArgs, "and"))
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
	invalidResourceSuggestion := fmt.Sprintf("List available %s with `%s list`.", plural.Plural(resourceType), fullParentCommand)

	return errors.NewErrorWithSuggestions(invalidArgsErrMsg, invalidResourceSuggestion)
}
