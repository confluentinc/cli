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
	Environment                     = "environment"
	FlinkComputePool                = "Flink compute pool"
	FlinkRegion                     = "Flink region"
	FlinkStatement                  = "Flink SQL statement"
	IdentityPool                    = "identity pool"
	IdentityProvider                = "identity provider"
	KafkaCluster                    = "Kafka cluster"
	KsqlCluster                     = "KSQL cluster"
	MirrorTopic                     = "mirror topic"
	Network                         = "network"
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
	EnvironmentPrefix:           Environment,
	IdentityPoolPrefix:          IdentityPool,
	IdentityProviderPrefix:      IdentityProvider,
	KafkaClusterPrefix:          KafkaCluster,
	KsqlClusterPrefix:           KsqlCluster,
	SchemaRegistryClusterPrefix: SchemaRegistryCluster,
	ServiceAccountPrefix:        ServiceAccount,
	UserPrefix:                  User,
}

var resourceToPrefix = map[string]string{
	Environment:           EnvironmentPrefix,
	IdentityPool:          IdentityPoolPrefix,
	IdentityProvider:      IdentityProviderPrefix,
	KafkaCluster:          KafkaClusterPrefix,
	KsqlCluster:           KsqlClusterPrefix,
	SchemaRegistryCluster: SchemaRegistryClusterPrefix,
	ServiceAccount:        ServiceAccountPrefix,
	User:                  UserPrefix,
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

func ValidatePrefixes(resourceType string, args []string) error {
	prefix, ok := resourceToPrefix[resourceType]
	if !ok {
		return nil
	}

	var malformed []string
	for _, resourceId := range args {
		if LookupType(resourceId) != resourceType {
			malformed = append(malformed, resourceId)
		}
	}

	if len(malformed) == 1 {
		return errors.Errorf(`failed parsing resource ID %s: missing prefix "%s-"`, malformed[0], prefix)
	} else if len(malformed) > 1 {
		return errors.Errorf(`failed parsing resource IDs %s: missing prefix "%s-"`, utils.ArrayToCommaDelimitedString(malformed, "and"), prefix)
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
