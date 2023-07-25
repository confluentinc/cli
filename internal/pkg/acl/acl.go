package acl

import (
	"fmt"
	"sort"
	"strings"

	"github.com/antihax/optional"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	cckafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	cpkafkarestv3 "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var listFields = []string{"Principal", "Permission", "Operation", "ResourceType", "ResourceName", "PatternType"}

var AclOperations = []mdsv1.AclOperation{
	mdsv1.ACLOPERATION_ALL,
	mdsv1.ACLOPERATION_ALTER,
	mdsv1.ACLOPERATION_ALTER_CONFIGS,
	mdsv1.ACLOPERATION_CLUSTER_ACTION,
	mdsv1.ACLOPERATION_CREATE,
	mdsv1.ACLOPERATION_DELETE,
	mdsv1.ACLOPERATION_DESCRIBE,
	mdsv1.ACLOPERATION_DESCRIBE_CONFIGS,
	mdsv1.ACLOPERATION_IDEMPOTENT_WRITE,
	mdsv1.ACLOPERATION_READ,
	mdsv1.ACLOPERATION_WRITE,
}

type out struct {
	Principal    string `human:"Principal" serialized:"principal"`
	Permission   string `human:"Permission" serialized:"permission"`
	Operation    string `human:"Operation" serialized:"operation"`
	Host         string `human:"Host" serialized:"host"`
	ResourceType string `human:"Resource Type" serialized:"resource_type"`
	ResourceName string `human:"Resource Name" serialized:"resource_name"`
	PatternType  string `human:"Pattern Type" serialized:"pattern_type"`
}

type AclRequestDataWithError struct {
	ResourceType cpkafkarestv3.AclResourceType
	ResourceName string
	PatternType  string
	Principal    string
	Host         string
	Operation    string
	Permission   string
	Errors       error
}

func PrintACLsFromKafkaRestResponseOnPrem(cmd *cobra.Command, acls []cpkafkarestv3.AclData) error {
	list := output.NewList(cmd)
	for _, acl := range acls {
		list.Add(&out{
			Principal:    acl.Principal,
			Permission:   acl.Permission,
			Operation:    acl.Operation,
			Host:         acl.Host,
			ResourceType: string(acl.ResourceType),
			ResourceName: acl.ResourceName,
			PatternType:  acl.PatternType,
		})
	}
	return list.Print()
}

func PrintACLs(cmd *cobra.Command, acls []*ccstructs.ACLBinding) error {
	list := output.NewList(cmd)
	for _, acl := range acls {
		list.Add(&out{
			Principal:    acl.GetEntry().GetPrincipal(),
			Permission:   acl.GetEntry().GetPermissionType().String(),
			Operation:    acl.GetEntry().GetOperation().String(),
			ResourceType: acl.GetPattern().GetResourceType().String(),
			ResourceName: acl.GetPattern().GetName(),
			PatternType:  acl.GetPattern().GetPatternType().String(),
		})
	}
	list.Filter(listFields)
	return list.Print()
}

func AclFlags() *pflag.FlagSet {
	flgSet := pflag.NewFlagSet("acl-config", pflag.ExitOnError)
	flgSet.String("principal", "", "Principal for this operation with User: or Group: prefix.")
	flgSet.String("operation", "", fmt.Sprintf("Set ACL Operation to: (%s).", ConvertToLower(AclOperations)))
	flgSet.String("host", "*", "Set host for access. Only IP addresses are supported.")
	flgSet.Bool("allow", false, "ACL permission to allow access.")
	flgSet.Bool("deny", false, "ACL permission to restrict access to resource.")
	flgSet.Bool("cluster-scope", false, "Set the cluster resource. With this option the ACL grants access to the provided operations on the Kafka cluster itself.")
	flgSet.String("consumer-group", "", "Set the Consumer Group resource.")
	flgSet.String("transactional-id", "", "Set the TransactionalID resource.")
	flgSet.String("topic", "", "Set the topic resource. With this option the ACL grants the provided operations on the topics that start with that prefix, depending on whether the --prefix option was also passed.")
	flgSet.Bool("prefix", false, "Set to match all resource names prefixed with this value.")
	flgSet.SortFlags = false
	return flgSet
}

func ParseAclRequest(cmd *cobra.Command) *AclRequestDataWithError {
	aclRequest := &AclRequestDataWithError{
		Host:   "*",
		Errors: nil,
	}
	cmd.Flags().Visit(populateAclRequest(aclRequest))
	return aclRequest
}

func populateAclRequest(conf *AclRequestDataWithError) func(*pflag.Flag) {
	return func(flag *pflag.Flag) {
		v := flag.Value.String()
		switch n := flag.Name; n {
		case "consumer-group":
			setAclRequestResourcePattern(conf, "GROUP", v) // set aclRequestData.ResourceType and aclRequestData.ResourceName
		case "cluster-scope":
			// The only valid name for a cluster is kafka-cluster
			// https://github.com/confluentinc/cc-kafka/blob/88823c6016ea2e306340938994d9e122abf3c6c0/core/src/main/scala/kafka/security/auth/Resource.scala#L24
			setAclRequestResourcePattern(conf, "cluster", "kafka-cluster")
		case "topic":
			fallthrough
		case "delegation-token":
			fallthrough
		case "transactional-id":
			setAclRequestResourcePattern(conf, n, v)
		case "allow":
			setAclRequestPermission(conf, "ALLOW")
		case "deny":
			setAclRequestPermission(conf, "DENY")
		case "prefix":
			conf.PatternType = "PREFIXED"
		case "principal":
			conf.Principal = v
		case "host":
			conf.Host = v
		case "operation":
			v = strings.ToUpper(v)
			v = strings.ReplaceAll(v, "-", "_")
			enumUtils := utils.EnumUtils{}
			enumUtils.Init(
				"UNKNOWN",
				"ANY",
				"ALL",
				"READ",
				"WRITE",
				"CREATE",
				"DELETE",
				"ALTER",
				"DESCRIBE",
				"CLUSTER_ACTION",
				"DESCRIBE_CONFIGS",
				"ALTER_CONFIGS",
				"IDEMPOTENT_WRITE",
			)
			if op, ok := enumUtils[v]; ok {
				conf.Operation = op.(string)
				break
			}
			conf.Errors = multierror.Append(conf.Errors, fmt.Errorf("invalid operation value: %s", v))
		}
	}
}

func setAclRequestPermission(conf *AclRequestDataWithError, permission string) {
	if conf.Permission != "" {
		conf.Errors = multierror.Append(conf.Errors, errors.Errorf(errors.OnlySetAllowOrDenyErrorMsg))
	}
	conf.Permission = permission
}

func setAclRequestResourcePattern(conf *AclRequestDataWithError, n, v string) {
	if conf.ResourceType != "" {
		// A resourceType has already been set with a previous flag
		conf.Errors = multierror.Append(conf.Errors, fmt.Errorf("exactly one of %v must be set",
			convertToFlags(cpkafkarestv3.ACLRESOURCETYPE_TOPIC, cpkafkarestv3.ACLRESOURCETYPE_GROUP,
				cpkafkarestv3.ACLRESOURCETYPE_CLUSTER, cpkafkarestv3.ACLRESOURCETYPE_TRANSACTIONAL_ID)))
		return
	}

	// Normalize the resource pattern name
	n = strings.ToUpper(n)
	n = strings.ReplaceAll(n, "-", "_")

	enumUtils := utils.EnumUtils{}
	enumUtils.Init(cpkafkarestv3.ACLRESOURCETYPE_TOPIC, cpkafkarestv3.ACLRESOURCETYPE_GROUP,
		cpkafkarestv3.ACLRESOURCETYPE_CLUSTER, cpkafkarestv3.ACLRESOURCETYPE_TRANSACTIONAL_ID)
	conf.ResourceType = enumUtils[n].(cpkafkarestv3.AclResourceType)

	if conf.ResourceType == cpkafkarestv3.ACLRESOURCETYPE_CLUSTER {
		conf.PatternType = "LITERAL"
	}
	conf.ResourceName = v
}

func convertToFlags(operations ...any) string {
	ops := make([]string, len(operations))

	for i, v := range operations {
		// clean the resources that don't map directly to flag name
		if v == cpkafkarestv3.ACLRESOURCETYPE_GROUP {
			v = "consumer-group"
		}
		if v == cpkafkarestv3.ACLRESOURCETYPE_CLUSTER {
			v = "cluster-scope"
		}
		s := strings.ToLower(strings.ReplaceAll(fmt.Sprint(v), "_", "-"))
		ops[i] = fmt.Sprintf("`--%s`", s)
	}

	sort.Strings(ops)
	return strings.Join(ops, ", ")
}

func ConvertToLower[T any](operations []T) string {
	ops := make([]string, len(operations))

	for i, v := range operations {
		ops[i] = strings.ReplaceAll(fmt.Sprint(v), "_", "-")
	}

	return strings.ToLower(strings.Join(ops, ", "))
}

func ValidateCreateDeleteAclRequestData(aclConfiguration *AclRequestDataWithError) *AclRequestDataWithError {
	// delete is deliberately less powerful in the cli than in the API to prevent accidental
	// deletion of too many acls at once. Expectation is that multi delete will be done via
	// repeated invocation of the cli by external scripts.
	if aclConfiguration.Permission == "" {
		aclConfiguration.Errors = multierror.Append(aclConfiguration.Errors, errors.Errorf(errors.MustSetAllowOrDenyErrorMsg))
	}

	if aclConfiguration.PatternType == "" {
		aclConfiguration.PatternType = "LITERAL"
	}

	if aclConfiguration.ResourceType == "" {
		aclConfiguration.Errors = multierror.Append(aclConfiguration.Errors, errors.Errorf(errors.MustSetResourceTypeErrorMsg,
			convertToFlags(cpkafkarestv3.ACLRESOURCETYPE_TOPIC, cpkafkarestv3.ACLRESOURCETYPE_GROUP,
				cpkafkarestv3.ACLRESOURCETYPE_CLUSTER, cpkafkarestv3.ACLRESOURCETYPE_TRANSACTIONAL_ID)))
	}
	return aclConfiguration
}

func AclRequestToCreateAclRequest(acl *AclRequestDataWithError) *cpkafkarestv3.CreateKafkaAclsOpts {
	return &cpkafkarestv3.CreateKafkaAclsOpts{
		CreateAclRequestData: optional.NewInterface(cpkafkarestv3.CreateAclRequestData{
			ResourceType: acl.ResourceType,
			ResourceName: acl.ResourceName,
			PatternType:  acl.PatternType,
			Principal:    acl.Principal,
			Host:         acl.Host,
			Operation:    acl.Operation,
			Permission:   acl.Permission,
		}),
	}
}

// Functions for converting AclRequestDataWithError into structs for create, delete, and list requests

func AclRequestToListAclRequest(acl *AclRequestDataWithError) *cpkafkarestv3.GetKafkaAclsOpts {
	opts := &cpkafkarestv3.GetKafkaAclsOpts{}
	if acl.ResourceType != "" {
		opts.ResourceType = optional.NewInterface(acl.ResourceType)
	}
	if acl.ResourceName != "" {
		opts.ResourceName = optional.NewString(acl.ResourceName)
	}
	if acl.PatternType != "" {
		opts.PatternType = optional.NewString(acl.PatternType)
	}
	if acl.Principal != "" {
		opts.Principal = optional.NewString(acl.Principal)
	}
	if acl.Host != "" {
		opts.Host = optional.NewString(acl.Host)
	}
	if acl.Operation != "" {
		opts.Operation = optional.NewString(acl.Operation)
	}
	if acl.Permission != "" {
		opts.Permission = optional.NewString(acl.Permission)
	}
	return opts
}

func AclRequestToDeleteAclRequest(acl *AclRequestDataWithError) *cpkafkarestv3.DeleteKafkaAclsOpts {
	return &cpkafkarestv3.DeleteKafkaAclsOpts{
		ResourceType: optional.NewInterface(acl.ResourceType),
		ResourceName: optional.NewString(acl.ResourceName),
		PatternType:  optional.NewString(acl.PatternType),
		Principal:    optional.NewString(acl.Principal),
		Host:         optional.NewString(acl.Host),
		Operation:    optional.NewString(acl.Operation),
		Permission:   optional.NewString(acl.Permission),
	}
}

func CreateAclRequestDataToAclData(data *AclRequestDataWithError) cpkafkarestv3.AclData {
	return cpkafkarestv3.AclData{
		ResourceType: data.ResourceType,
		ResourceName: data.ResourceName,
		PatternType:  data.PatternType,
		Principal:    data.Principal,
		Host:         data.Host,
		Operation:    data.Operation,
		Permission:   data.Permission,
	}
}

func PrintACLsFromKafkaRestResponse(cmd *cobra.Command, acls []cckafkarestv3.AclData) error {
	list := output.NewList(cmd)
	for _, acl := range acls {
		list.Add(&out{
			Principal:    acl.GetPrincipal(),
			Permission:   acl.GetPermission(),
			Operation:    acl.GetOperation(),
			ResourceType: string(acl.GetResourceType()),
			ResourceName: acl.GetResourceName(),
			PatternType:  acl.GetPatternType(),
		})
	}
	list.Filter(listFields)
	return list.Print()
}

func GetCreateAclRequestData(binding *ccstructs.ACLBinding) cckafkarestv3.CreateAclRequestData {
	data := cckafkarestv3.CreateAclRequestData{
		Host:         binding.GetEntry().GetHost(),
		Principal:    binding.GetEntry().GetPrincipal(),
		ResourceName: binding.GetPattern().GetName(),
	}

	if binding.GetPattern().GetResourceType() != ccstructs.ResourceTypes_UNKNOWN {
		data.ResourceType = cckafkarestv3.AclResourceType(binding.GetPattern().GetResourceType().String())
	}

	if binding.GetPattern().GetPatternType() != ccstructs.PatternTypes_UNKNOWN {
		data.PatternType = binding.GetPattern().GetPatternType().String()
	}

	if binding.GetEntry().GetOperation() != ccstructs.ACLOperations_UNKNOWN {
		data.Operation = binding.GetEntry().GetOperation().String()
	}

	if binding.GetEntry().GetPermissionType() != ccstructs.ACLPermissionTypes_UNKNOWN {
		data.Permission = binding.GetEntry().GetPermissionType().String()
	}

	return data
}
