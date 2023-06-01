package resource

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
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

	if last := string(resource[len(resource)-1]); last == "s" || last == "x" || last == "z" {
		return resource + "es"
	}
	if len(resource) > 1 {
		if lastTwo := resource[len(resource)-2:]; lastTwo == "ch" || lastTwo == "sh" {
			return resource + "es"
		}
	}

	return resource + "s"
}

func DefaultPostProcess(_ string) error {
	return nil
}

func Delete(args []string, callDeleteEndpoint func(string) error, postProcess func(string) error) ([]string, error) {
	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	var deleted []string
	for _, id := range args {
		if err := callDeleteEndpoint(id); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			deleted = append(deleted, id)
			if err := postProcess(id); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}

	return deleted, errs.ErrorOrNil()
}

func PrintDeleteSuccessMsg(successful []string, resourceType string) {
	if len(successful) == 1 {
		output.Printf(errors.DeletedResourceMsg, resourceType, successful[0])
	} else if len(successful) > 1 {
		output.Printf("Deleted %s %s.\n", Plural(resourceType), utils.ArrayToCommaDelimitedString(successful, "and"))
	}
}
