package acl

import (
	"fmt"
	"github.com/antihax/optional"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	krsdk "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"io"
	"sort"
	"strings"
)

type AclRequestDataWithError struct {
	ResourceType kafkarestv3.AclResourceType
	ResourceName string
	PatternType  kafkarestv3.AclPatternType
	Principal    string
	Host         string
	Operation    kafkarestv3.AclOperation
	Permission   kafkarestv3.AclPermission
	Errors error
}

type enumUtils map[string]interface{}

func (enumUtils enumUtils) init(enums ...interface{}) enumUtils {
	for _, enum := range enums {
		enumUtils[fmt.Sprintf("%v", enum)] = enum
	}
	return enumUtils
}

func PrintACLsFromKafkaRestResponse(cmd *cobra.Command, aclGetResp []krsdk.AclData, writer io.Writer, aclListFields, aclListStructuredRenames []string) error {
	// non list commands which do not have -o flags also uses this function, need to set default
	_, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	}
	outputWriter, err := output.NewListOutputCustomizableWriter(cmd, aclListFields, aclListFields, aclListStructuredRenames, writer)
	if err != nil {
		return err
	}

	for _, aclData := range aclGetResp {
		record := &struct {
			ServiceAccountId string
			Principal		 string
			Permission       string
			Operation        string
			Host			 string
			Resource         string
			ResourceType	 string
			Name             string
			ResourceName	 string
			Type             string
			PatternType		 string
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
		cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	}

	aclListFields := []string{"ServiceAccountId", "Permission", "Operation", "Resource", "Name", "Type"}
	aclListStructuredRenames := []string{"service_account_id", "permission", "operation", "resource", "name", "type"}
	outputWriter, err := output.NewListOutputCustomizableWriter(cmd, aclListFields, aclListFields, aclListStructuredRenames, writer)
	if err != nil {
		return err
	}

	for _, binding := range bindingsObj {
		record := &struct {
			ServiceAccountId string
			Permission       string
			Operation        string
			Resource         string
			Name             string
			Type             string
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
			conf.Permission = kafkarestv3.ACLPERMISSION_ALLOW
		case "deny":
			conf.Permission = kafkarestv3.ACLPERMISSION_DENY
		case "prefix":
			conf.PatternType = kafkarestv3.ACLPATTERNTYPE_PREFIXED
		case "principal":
			conf.Principal = v
		case "host":
			conf.Host = v
		case "operation":
			v = strings.ToUpper(v)
			v = strings.ReplaceAll(v, "-", "_")
			enumUtils := enumUtils{}
			enumUtils.init(
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

func setAclRequestResourcePattern(conf *AclRequestDataWithError, n, v string) {
	if conf.ResourceType != "" {
		// A resourceType has already been set with a previous flag
		conf.Errors = multierror.Append(conf.Errors, fmt.Errorf("exactly one of %v must be set",
			ConvertToFlags(kafkarestv3.ACLRESOURCETYPE_TOPIC, kafkarestv3.ACLRESOURCETYPE_GROUP,
				kafkarestv3.ACLRESOURCETYPE_CLUSTER, kafkarestv3.ACLRESOURCETYPE_TRANSACTIONAL_ID)))
		return
	}

	// Normalize the resource pattern name
	n = strings.ToUpper(n)
	n = strings.ReplaceAll(n, "-", "_")

	enumUtils := enumUtils{}
	enumUtils.init(kafkarestv3.ACLRESOURCETYPE_TOPIC, kafkarestv3.ACLRESOURCETYPE_GROUP,
		kafkarestv3.ACLRESOURCETYPE_CLUSTER, kafkarestv3.ACLRESOURCETYPE_TRANSACTIONAL_ID)
	conf.ResourceType = enumUtils[n].(kafkarestv3.AclResourceType)

	if conf.ResourceType == kafkarestv3.ACLRESOURCETYPE_CLUSTER {
		conf.PatternType = kafkarestv3.ACLPATTERNTYPE_LITERAL
	}
	conf.ResourceName = v
}

func ConvertToFlags(operations ...interface{}) string {
	var ops []string

	for _, v := range operations {
		if v == krsdk.ACLRESOURCETYPE_GROUP {
			v = "consumer-group"
		}
		if v == krsdk.ACLRESOURCETYPE_CLUSTER {
			v = "cluster-scope"
		}
		s := fmt.Sprintf("%v", v)
		s = strings.ReplaceAll(s, "_", "-")
		ops = append(ops, strings.ToLower(s))
	}

	sort.Strings(ops)
	return strings.Join(ops, ", ")
}

func AclRequestToCreateAclReqest(acl *AclRequestDataWithError) *krsdk.ClustersClusterIdAclsPostOpts {
	var opts krsdk.ClustersClusterIdAclsPostOpts
	requestData := krsdk.CreateAclRequestData{
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

func AclRequestToListAclReqest(acl *AclRequestDataWithError) *krsdk.ClustersClusterIdAclsGetOpts {
	opts := krsdk.ClustersClusterIdAclsGetOpts{
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

func AclRequestToDeleteAclReqest(acl *AclRequestDataWithError) *krsdk.ClustersClusterIdAclsDeleteOpts {
	opts := krsdk.ClustersClusterIdAclsDeleteOpts{
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

func CreateAclRequestDataToAclData(data *AclRequestDataWithError) krsdk.AclData {
	aclData := krsdk.AclData{
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
