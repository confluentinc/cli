package acl

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/antihax/optional"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type AclRequestDataWithError struct {
	ResourceType kafkarestv3.AclResourceType
	ResourceName string
	PatternType  kafkarestv3.AclPatternType
	Principal    string
	Host         string
	Operation    kafkarestv3.AclOperation
	Permission   kafkarestv3.AclPermission
	Errors       error
}

func PrintACLsFromKafkaRestResponse(cmd *cobra.Command, aclGetResp []kafkarestv3.AclData, writer io.Writer, aclListFields, aclListStructuredRenames []string) error {
	// non list commands which do not have -o flags also uses this function, need to set default
	_, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		pcmd.AddOutputFlag(cmd)
	}
	outputWriter, err := output.NewListOutputCustomizableWriter(cmd, aclListFields, aclListFields, aclListStructuredRenames, writer)
	if err != nil {
		return err
	}

	for _, aclData := range aclGetResp {
		record := &struct { //TODO remove KafkaAPI field names and move to only Kafka REST ones
			ServiceAccountId string
			Principal        string
			Permission       string
			Operation        string
			Host             string
			Resource         string
			ResourceType     string
			Name             string
			ResourceName     string
			Type             string
			PatternType      string
		}{
			aclData.Principal,
			aclData.Principal,
			string(aclData.Permission),
			string(aclData.Operation),
			aclData.Host,
			string(aclData.ResourceType),
			string(aclData.ResourceType),
			aclData.ResourceName,
			aclData.ResourceName,
			string(aclData.PatternType),
			string(aclData.PatternType),
		}
		outputWriter.AddElement(record)
	}

	return outputWriter.Out()
}

func PrintACLs(cmd *cobra.Command, bindingsObj []*schedv1.ACLBinding, writer io.Writer) error {
	// non list commands which do not have -o flags also uses this function, need to set default
	_, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		pcmd.AddOutputFlag(cmd)
	}

	aclListFields := []string{"Principal", "Permission", "Operation", "ResourceType", "ResourceName", "PatternType"}
	aclListStructuredRenames := []string{"principal", "permission", "operation", "resource_type", "resource_name", "pattern_type"}
	outputWriter, err := output.NewListOutputCustomizableWriter(cmd, aclListFields, aclListFields, aclListStructuredRenames, writer)
	if err != nil {
		return err
	}

	for _, binding := range bindingsObj {
		record := &struct {
			Principal    string
			Permission   string
			Operation    string
			ResourceType string
			ResourceName string
			PatternType  string
		}{
			binding.Entry.Principal,
			binding.Entry.PermissionType.String(),
			binding.Entry.Operation.String(),
			binding.Pattern.ResourceType.String(),
			binding.Pattern.Name,
			binding.Pattern.PatternType.String(),
		}
		outputWriter.AddElement(record)
	}

	return outputWriter.Out()
}

func CreateACLFlags() *pflag.FlagSet {
	flgSet := AclFlags()
	_ = cobra.MarkFlagRequired(flgSet, "principal")
	_ = cobra.MarkFlagRequired(flgSet, "operation")
	return flgSet
}

func DeleteACLFlags() *pflag.FlagSet {
	flgSet := AclFlags()
	_ = cobra.MarkFlagRequired(flgSet, "principal")
	_ = cobra.MarkFlagRequired(flgSet, "operation")
	_ = cobra.MarkFlagRequired(flgSet, "host")
	return flgSet
}

func AclFlags() *pflag.FlagSet {
	flgSet := pflag.NewFlagSet("acl-config", pflag.ExitOnError)
	flgSet.Bool("allow", false, "ACL permission to allow access.")
	flgSet.Bool("deny", false, "ACL permission to restrict access to resource.")
	flgSet.String("principal", "", "Principal for this operation with User: or Group: prefix.")
	flgSet.String("host", "*", "Set host for access. Only IP addresses are supported.")
	flgSet.String("operation", "", fmt.Sprintf("Set ACL Operation to: (%s).",
		convertToFlags(kafkarestv3.ACLOPERATION_ALL, kafkarestv3.ACLOPERATION_READ, kafkarestv3.ACLOPERATION_WRITE,
			kafkarestv3.ACLOPERATION_CREATE, kafkarestv3.ACLOPERATION_DELETE, kafkarestv3.ACLOPERATION_ALTER,
			kafkarestv3.ACLOPERATION_DESCRIBE, kafkarestv3.ACLOPERATION_CLUSTER_ACTION,
			kafkarestv3.ACLOPERATION_DESCRIBE_CONFIGS, kafkarestv3.ACLOPERATION_ALTER_CONFIGS,
			kafkarestv3.ACLOPERATION_IDEMPOTENT_WRITE)))
	flgSet.Bool("cluster-scope", false, `Set the cluster resource. With this option the ACL grants
access to the provided operations on the Kafka cluster itself.`)
	flgSet.String("consumer-group", "", "Set the Consumer Group resource.")
	flgSet.String("transactional-id", "", "Set the TransactionalID resource.")
	flgSet.String("topic", "", `Set the topic resource. With this option the ACL grants the provided
operations on the topics that start with that prefix, depending on whether
the --prefix option was also passed.`)
	flgSet.Bool("prefix", false, "Set to match all resource names prefixed with this value.")
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
			setAclRequestPermission(conf, kafkarestv3.ACLPERMISSION_ALLOW)
		case "deny":
			setAclRequestPermission(conf, kafkarestv3.ACLPERMISSION_DENY)
		case "prefix":
			conf.PatternType = kafkarestv3.ACLPATTERNTYPE_PREFIXED
		case "principal":
			conf.Principal = v
		case "host":
			conf.Host = v
		case "operation":
			v = strings.ToUpper(v)
			v = strings.ReplaceAll(v, "-", "_")
			enumUtils := utils.EnumUtils{}
			enumUtils.Init(
				kafkarestv3.ACLOPERATION_UNKNOWN,
				kafkarestv3.ACLOPERATION_ANY,
				kafkarestv3.ACLOPERATION_ALL,
				kafkarestv3.ACLOPERATION_READ,
				kafkarestv3.ACLOPERATION_WRITE,
				kafkarestv3.ACLOPERATION_CREATE,
				kafkarestv3.ACLOPERATION_DELETE,
				kafkarestv3.ACLOPERATION_ALTER,
				kafkarestv3.ACLOPERATION_DESCRIBE,
				kafkarestv3.ACLOPERATION_CLUSTER_ACTION,
				kafkarestv3.ACLOPERATION_DESCRIBE_CONFIGS,
				kafkarestv3.ACLOPERATION_ALTER_CONFIGS,
				kafkarestv3.ACLOPERATION_IDEMPOTENT_WRITE,
			)
			if op, ok := enumUtils[v]; ok {
				conf.Operation = op.(kafkarestv3.AclOperation)
				break
			}
			conf.Errors = multierror.Append(conf.Errors, fmt.Errorf("Invalid operation value: "+v))
		}
	}
}

func setAclRequestPermission(conf *AclRequestDataWithError, permission kafkarestv3.AclPermission) {
	if conf.Permission != "" {
		conf.Errors = multierror.Append(conf.Errors, errors.Errorf(errors.OnlySetAllowOrDenyErrorMsg))
	}
	conf.Permission = permission
}

func setAclRequestResourcePattern(conf *AclRequestDataWithError, n, v string) {
	if conf.ResourceType != "" {
		// A resourceType has already been set with a previous flag
		conf.Errors = multierror.Append(conf.Errors, fmt.Errorf("exactly one of %v must be set",
			convertToFlags(kafkarestv3.ACLRESOURCETYPE_TOPIC, kafkarestv3.ACLRESOURCETYPE_GROUP,
				kafkarestv3.ACLRESOURCETYPE_CLUSTER, kafkarestv3.ACLRESOURCETYPE_TRANSACTIONAL_ID)))
		return
	}

	// Normalize the resource pattern name
	n = strings.ToUpper(n)
	n = strings.ReplaceAll(n, "-", "_")

	enumUtils := utils.EnumUtils{}
	enumUtils.Init(kafkarestv3.ACLRESOURCETYPE_TOPIC, kafkarestv3.ACLRESOURCETYPE_GROUP,
		kafkarestv3.ACLRESOURCETYPE_CLUSTER, kafkarestv3.ACLRESOURCETYPE_TRANSACTIONAL_ID)
	conf.ResourceType = enumUtils[n].(kafkarestv3.AclResourceType)

	if conf.ResourceType == kafkarestv3.ACLRESOURCETYPE_CLUSTER {
		conf.PatternType = kafkarestv3.ACLPATTERNTYPE_LITERAL
	}
	conf.ResourceName = v
}

func convertToFlags(operations ...interface{}) string {
	var ops []string

	for _, v := range operations {
		// clean the resources that don't map directly to flag name
		if v == kafkarestv3.ACLRESOURCETYPE_GROUP {
			v = "consumer-group"
		}
		if v == kafkarestv3.ACLRESOURCETYPE_CLUSTER {
			v = "cluster-scope"
		}
		s := fmt.Sprintf("%v", v)
		s = strings.ReplaceAll(s, "_", "-")
		ops = append(ops, strings.ToLower(s))
	}

	sort.Strings(ops)
	return strings.Join(ops, ", ")
}

func ValidateCreateDeleteAclRequestData(aclConfiguration *AclRequestDataWithError) *AclRequestDataWithError {
	// delete is deliberately less powerful in the cli than in the API to prevent accidental
	// deletion of too many acls at once. Expectation is that multi delete will be done via
	// repeated invocation of the cli by external scripts.
	if aclConfiguration.Permission == "" {
		aclConfiguration.Errors = multierror.Append(aclConfiguration.Errors, errors.Errorf(errors.MustSetAllowOrDenyErrorMsg))
	}

	if aclConfiguration.PatternType == "" {
		aclConfiguration.PatternType = kafkarestv3.ACLPATTERNTYPE_LITERAL
	}

	if aclConfiguration.ResourceType == "" {
		aclConfiguration.Errors = multierror.Append(aclConfiguration.Errors, errors.Errorf(errors.MustSetResourceTypeErrorMsg,
			convertToFlags(kafkarestv3.ACLRESOURCETYPE_TOPIC, kafkarestv3.ACLRESOURCETYPE_GROUP,
				kafkarestv3.ACLRESOURCETYPE_CLUSTER, kafkarestv3.ACLRESOURCETYPE_TRANSACTIONAL_ID)))
	}
	return aclConfiguration
}

func AclRequestToCreateAclReqest(acl *AclRequestDataWithError) *kafkarestv3.ClustersClusterIdAclsPostOpts {
	var opts kafkarestv3.ClustersClusterIdAclsPostOpts
	requestData := kafkarestv3.CreateAclRequestData{
		ResourceType: acl.ResourceType,
		ResourceName: acl.ResourceName,
		PatternType:  acl.PatternType,
		Principal:    acl.Principal,
		Host:         acl.Host,
		Operation:    acl.Operation,
		Permission:   acl.Permission,
	}
	opts.CreateAclRequestData = optional.NewInterface(requestData)
	return &opts
}

// Functions for converting AclRequestDataWithError into structs for create, delete, and list requests

func AclRequestToListAclReqest(acl *AclRequestDataWithError) *kafkarestv3.ClustersClusterIdAclsGetOpts {
	opts := kafkarestv3.ClustersClusterIdAclsGetOpts{}
	if acl.ResourceType != "" {
		opts.ResourceType = optional.NewInterface(acl.ResourceType)
	}
	if acl.ResourceName != "" {
		opts.ResourceName = optional.NewString(acl.ResourceName)
	}
	if acl.PatternType != "" {
		opts.PatternType = optional.NewInterface(acl.PatternType)
	}
	if acl.Principal != "" {
		opts.Principal = optional.NewString(acl.Principal)
	}
	if acl.Host != "" {
		opts.Host = optional.NewString(acl.Host)
	}
	if acl.Operation != "" {
		opts.Operation = optional.NewInterface(acl.Operation)
	}
	if acl.Permission != "" {
		opts.Permission = optional.NewInterface(acl.Permission)
	}
	return &opts
}

func AclRequestToDeleteAclReqest(acl *AclRequestDataWithError) *kafkarestv3.ClustersClusterIdAclsDeleteOpts {
	opts := kafkarestv3.ClustersClusterIdAclsDeleteOpts{
		ResourceType: optional.NewInterface(acl.ResourceType),
		ResourceName: optional.NewString(acl.ResourceName),
		PatternType:  optional.NewInterface(acl.PatternType),
		Principal:    optional.NewString(acl.Principal),
		Host:         optional.NewString(acl.Host),
		Operation:    optional.NewInterface(acl.Operation),
		Permission:   optional.NewInterface(acl.Permission),
	}
	return &opts
}

func CreateAclRequestDataToAclData(data *AclRequestDataWithError) kafkarestv3.AclData {
	aclData := kafkarestv3.AclData{
		ResourceType: data.ResourceType,
		ResourceName: data.ResourceName,
		PatternType:  data.PatternType,
		Principal:    data.Principal,
		Host:         data.Host,
		Operation:    data.Operation,
		Permission:   data.Permission,
	}
	return aclData
}

func PrintACLsFromKafkaRestResponseWithResourceIdMap(cmd *cobra.Command, aclGetResp kafkarestv3.AclDataList, writer io.Writer, idMap map[int32]string) error {
	// non list commands which do not have -o flags also uses this function, need to set default
	_, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		pcmd.AddOutputFlag(cmd)
	}

	aclListFields := []string{"Principal", "Permission", "Operation", "ResourceType", "ResourceName", "PatternType"}
	aclListStructuredRenames := []string{"principal", "permission", "operation", "resource_type", "resource_name", "pattern_type"}
	outputWriter, err := output.NewListOutputCustomizableWriter(cmd, aclListFields, aclListFields, aclListStructuredRenames, writer)
	if err != nil {
		return err
	}

	for _, aclData := range aclGetResp.Data {
		principal := aclData.Principal
		prefix, resourceId, err := getPrefixAndResourceIdFromPrincipal(principal, idMap)
		if err != nil {
			if err.Error() == errors.UserNotFoundErrorMsg {
				continue // skip the entry if not a valid user id
			}
			return err
		}
		record := &struct {
			Principal    string
			Permission   string
			Operation    string
			ResourceType string
			ResourceName string
			PatternType  string
		}{
			prefix + ":" + resourceId,
			string(aclData.Permission),
			string(aclData.Operation),
			string(aclData.ResourceType),
			string(aclData.ResourceName),
			string(aclData.PatternType),
		}
		outputWriter.AddElement(record)
	}

	return outputWriter.Out()
}

func PrintACLsWithResourceIdMap(cmd *cobra.Command, bindingsObj []*schedv1.ACLBinding, writer io.Writer, idMap map[int32]string) error {
	// non list commands which do not have -o flags also uses this function, need to set default
	_, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		pcmd.AddOutputFlag(cmd)
	}

	aclListFields := []string{"Principal", "Permission", "Operation", "ResourceType", "ResourceName", "PatternType"}
	aclListStructuredRenames := []string{"principal", "permission", "operation", "resource_type", "resource_name", "pattern_type"}
	outputWriter, err := output.NewListOutputCustomizableWriter(cmd, aclListFields, aclListFields, aclListStructuredRenames, writer)
	if err != nil {
		return err
	}

	for _, binding := range bindingsObj {
		principal := binding.Entry.Principal
		prefix, resourceId, err := getPrefixAndResourceIdFromPrincipal(principal, idMap)
		if err != nil {
			if err.Error() == errors.UserNotFoundErrorMsg {
				continue // skip the entry if not a valid user id
			}
			return err
		}
		record := &struct {
			Principal    string
			Permission   string
			Operation    string
			ResourceType string
			ResourceName string
			PatternType  string
		}{
			prefix + ":" + resourceId,
			binding.Entry.PermissionType.String(),
			binding.Entry.Operation.String(),
			binding.Pattern.ResourceType.String(),
			binding.Pattern.Name,
			binding.Pattern.PatternType.String(),
		}
		outputWriter.AddElement(record)
	}

	return outputWriter.Out()
}

func getPrefixAndResourceIdFromPrincipal(principal string, idMap map[int32]string) (string, string, error) {
	var prefix, resourceId string
	if principal != "" {
		splitPrincipal := strings.Split(principal, ":")
		if len(splitPrincipal) < 2 {
			return prefix, resourceId, errors.Errorf("Unrecognized principal format %s", principal)
		}
		prefix = splitPrincipal[0]
		userId := splitPrincipal[1]
		idp, _ := strconv.ParseInt(userId, 10, 32)
		if _, ok := idMap[int32(idp)]; !ok {
			return "", "", errors.New(errors.UserNotFoundErrorMsg)
		}
		resourceId = idMap[int32(idp)]
	}
	return prefix, resourceId, nil
}
