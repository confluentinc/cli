package resource

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/types"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

const (
	Unknown                     = "unknown"
	ACL                         = "ACL"
	ApiKey                      = "API key"
	Broker                      = "broker"
	ByokKey                     = "self-managed key"
	ClientQuota                 = "client quota"
	Cloud                       = "cloud"
	ClusterLink                 = "cluster link"
	Connector                   = "connector"
	ConsumerShare               = "consumer share"
	Context                     = "context"
	Environment                 = "environment"
	FlinkComputePool            = "Flink compute pool"
	FlinkRegion                 = "Flink region"
	FlinkIamBinding             = "Flink IAM binding"
	FlinkStatement              = "Flink SQL statement"
	IdentityPool                = "identity pool"
	IdentityProvider            = "identity provider"
	KafkaCluster                = "Kafka cluster"
	KsqlCluster                 = "KSQL cluster"
	MirrorTopic                 = "mirror topic"
	Organization                = "organization"
	ProviderShare               = "provider share"
	Pipeline                    = "pipeline"
	SchemaExporter              = "schema exporter"
	SchemaRegistryCluster       = "Schema Registry cluster"
	SchemaRegistryConfiguration = "Schema Registry configuration"
	ServiceAccount              = "service account"
	SsoGroupMapping             = "SSO group mapping"
	Topic                       = "topic"
	User                        = "user"
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

var resourceToPrefix = map[string]string{
	ClusterLink:           ClusterLinkPrefix,
	Environment:           EnvironmentPrefix,
	IdentityPool:          IdentityPoolPrefix,
	IdentityProvider:      IdentityProviderPrefix,
	KafkaCluster:          KafkaClusterPrefix,
	KsqlCluster:           KsqlClusterPrefix,
	SchemaRegistryCluster: SchemaRegistryClusterPrefix,
	ServiceAccount:        ServiceAccountPrefix,
	User:                  UserPrefix,
}

// Singular words ending w/ these suffixes generally add an extra -es syllable in their plural forms
var pluralExtraSyllableSuffix = types.NewSet("s", "x", "z", "ch", "sh")

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

func ValidateArgs(fullParentCommand string, args []string, resourceType string, callDescribeEndpoint func(string) error) error {
	var invalidArgs []string
	for _, arg := range args {
		if err := callDescribeEndpoint(arg); err != nil {
			invalidArgs = append(invalidArgs, arg)
		}
	}

	if len(invalidArgs) != 0 {
		NotFoundErrorMsg := `%s %s not found`
		invalidArgsErrMsg := fmt.Sprintf(NotFoundErrorMsg, resourceType, utils.ArrayToCommaDelimitedString(invalidArgs, "and"))
		if len(invalidArgs) > 1 {
			invalidArgsErrMsg = fmt.Sprintf(NotFoundErrorMsg, Plural(resourceType), utils.ArrayToCommaDelimitedString(invalidArgs, "and"))
		}
		return errors.NewErrorWithSuggestions(invalidArgsErrMsg, fmt.Sprintf(errors.ListResourceSuggestions, resourceType, fullParentCommand))
	}

	return nil
}

func Plural(resource string) string {
	if resource == "" {
		return ""
	}

	for suffix := range pluralExtraSyllableSuffix {
		if strings.HasSuffix(resource, suffix) {
			return resource + "es"
		}
	}

	return resource + "s"
}

func delete(args []string, callDeleteEndpoint func(string) error) ([]string, error) {
	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	var deletedIDs []string
	for _, id := range args {
		if err := callDeleteEndpoint(id); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			deletedIDs = append(deletedIDs, id)
		}
	}

	return deletedIDs, errs.ErrorOrNil()
}

func Delete(args []string, callDeleteEndpoint func(string) error, resourceType string) ([]string, error) {
	deletedIDs, err := delete(args, callDeleteEndpoint)

	DeletedResourceMsg := "Deleted %s %s.\n"
	if len(deletedIDs) == 1 {
		output.Printf(DeletedResourceMsg, resourceType, fmt.Sprintf("\"%s\"", deletedIDs[0]))
	} else if len(deletedIDs) > 1 {
		output.Printf(DeletedResourceMsg, Plural(resourceType), utils.ArrayToCommaDelimitedString(deletedIDs, "and"))
	}

	return deletedIDs, err
}

func DeleteWithCustomMessage(args []string, callDeleteEndpoint func(string) error, singularMsg, pluralMsg string) ([]string, error) {
	deletedIDs, err := delete(args, callDeleteEndpoint)

	if len(deletedIDs) == 1 {
		output.Printf(singularMsg, fmt.Sprintf("\"%s\"", deletedIDs[0]))
	} else if len(deletedIDs) > 1 {
		output.Printf(pluralMsg, utils.ArrayToCommaDelimitedString(deletedIDs, "and"))
	}

	return deletedIDs, err
}
