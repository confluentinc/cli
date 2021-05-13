package kafka

import (
	"fmt"
	"github.com/antihax/optional"
	aclutil "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sort"
	"strings"
)

type aclOnPremCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func NewAclCommandOnPrem(prerunner pcmd.PreRunner) *cobra.Command {
	aclCmd := &aclOnPremCommand{
		pcmd.NewAuthenticatedStateFlagCommand(
			&cobra.Command{
				Use: "acl",
				Short: "Manage Kafka ACLs with Confluent Rest Proxy.",
			}, prerunner, OnPremTopicSubcommandFlags),
	}
	aclCmd.SetPersistentPreRunE(prerunner.InitializeOnPremKafkaRest(aclCmd.AuthenticatedCLICommand))
	aclCmd.init()
	return aclCmd.Command
}

func (aclCmd *aclOnPremCommand) init() {
	createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a Kafka ACL.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(aclCmd.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "You can specify only one of the following flags per command invocation: ``cluster``, ``consumer-group``, ``topic``, or ``transactional-id``. For example, to modify both ``consumer-group`` and ``topic`` resources, you need to issue two separate commands:",
				Code: "ccloud kafka acl create --allow --service-account 1522 --operation READ --consumer-group java_example_group_1\nccloud kafka acl create --allow --service-account 1522 --operation READ --topic '*'",
			}),
	}
	createCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	createCmd.Flags().AddFlagSet(aclFlags())
	//createCmd.Flags().Bool("allow", false, "Access to the resource is allowed.")
	//createCmd.Flags().Bool("deny", false, "Access to the resource is denied.")
	//createCmd.Flags().String("resource_type", "", "Resource type.")
	//createCmd.Flags().String("pattern_type", "", "Pattern type.")
	//createCmd.Flags().String("principal", "", "Principal.")
	//createCmd.Flags().String("operation", "", "Operation.")
//	createCmd.Flags().String("permission", "", "Permission.")
	createCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	createCmd.Flags().SortFlags = false
	aclCmd.AddCommand(createCmd)

	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a Kafka ACL.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(aclCmd.delete),
	}
	deleteCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	deleteCmd.Flags().AddFlagSet(aclConfigFlags())
	deleteCmd.Flags().SortFlags = false
	aclCmd.AddCommand(deleteCmd)

	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs for a resource.",
		Args: cobra.NoArgs,
		RunE: pcmd.NewCLIRunE(aclCmd.list),
	}
	listCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	listCmd.Flags().AddFlagSet(resourceFlags())
	listCmd.Flags().Int("service-account", 0, "Service account ID.")
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	aclCmd.AddCommand(listCmd)
}

func (aclCmd *aclOnPremCommand) list(cmd *cobra.Command, _ []string) error {
	acl, err := parse(cmd)
	if err != nil {
		return err
	}
	restClient, restContext, err := initKafkaRest(aclCmd.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	opts := aclBindingToClustersClusterIdAclsGetOpts(acl[0].ACLBinding)
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	aclGetResp, httpResp, err := restClient.ACLApi.ClustersClusterIdAclsGet(restContext, clusterId, &opts)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, httpResp)
	}
	return aclutil.PrintACLsFromKafkaRestResponse(cmd, aclGetResp, cmd.OutOrStdout())
}

//func (aclCmd *aclOnPremCommand) create(cmd *cobra.Command, _ []string) error {
//	acls, err := parse(cmd)
//	if err != nil {
//		return err
//	}
//	restClient, restContext, err := initKafkaRest(aclCmd.AuthenticatedCLICommand, cmd)
//	if err != nil {
//		return err
//	}
//	var bindings []*schedv1.ACLBinding
//	for _, acl := range acls {
//		validateAddAndDelete(acl)
//		if acl.errors != nil {
//			return acl.errors
//		}
//		bindings = append(bindings, acl.ACLBinding)
//	}
//	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
//	if err != nil {
//		return err
//	}
//	for i, binding := range bindings {
//		opts := aclBindingToClustersClusterIdAclsPostOpts(binding)
//		httpResp, err := restClient.ACLApi.ClustersClusterIdAclsPost(restContext, clusterId, &opts)
//		if err != nil {
//			if i > 0 {
//				_ = aclutil.PrintACLs(cmd, bindings[:i], cmd.OutOrStdout())
//			}
//			return kafkaRestError(restClient.GetConfig().BasePath, err, httpResp)
//		} else if httpResp != nil && httpResp.StatusCode != http.StatusCreated {
//			if i > 0 {
//				_ = aclutil.PrintACLs(cmd, bindings[:i], cmd.OutOrStdout())
//			}
//			return errors.NewErrorWithSuggestions(
//				fmt.Sprintf(errors.KafkaRestUnexpectedStatusMsg, httpResp.Request.URL, httpResp.StatusCode),
//				errors.InternalServerErrorSuggestions)
//		}
//	}
//	return aclutil.PrintACLs(cmd, bindings, cmd.OutOrStdout())
//}

func (aclCmd *aclOnPremCommand) create(cmd *cobra.Command, _ []string) error {
	acl := parseCreatAclRequest(cmd)
	if acl.errors != nil {
		return acl.errors
	}
	restClient, restContext, err := initKafkaRest(aclCmd.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	acl = validateCreateAclRequestData(acl)
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	var opts kafkarestv3.ClustersClusterIdAclsPostOpts
	opts.CreateAclRequestData = optional.NewInterface(*acl.CreateAclRequestData)
	httpResp, err := restClient.ACLApi.ClustersClusterIdAclsPost(restContext, clusterId, &opts)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, httpResp)
	}
	utils.Println(cmd, "yeeted a new acl")
	// TODO print the created acl
	return nil
}

func (aclCmd *aclOnPremCommand) delete(cmd *cobra.Command, _ []string) error {
	return nil
}

type CreateAclRequestDataWithError struct {
	*kafkarestv3.CreateAclRequestData
	errors error
}

func validateCreateAclRequestData(aclConfiguration *CreateAclRequestDataWithError) *CreateAclRequestDataWithError {
	// delete is deliberately less powerful in the cli than in the API to prevent accidental
	// deletion of too many acls at once. Expectation is that multi delete will be done via
	// repeated invocation of the cli by external scripts.
	if aclConfiguration.Permission == "" {
		aclConfiguration.errors = multierror.Append(aclConfiguration.errors, errors.Errorf(errors.MustSetAllowOrDenyErrorMsg))
	}

	if aclConfiguration.PatternType == "" {
		aclConfiguration.PatternType = kafkarestv3.ACLPATTERNTYPE_LITERAL
	}

	if aclConfiguration.ResourceType == "" {
		aclConfiguration.errors = multierror.Append(aclConfiguration.errors, errors.Errorf(errors.MustSetResourceTypeErrorMsg,
			convertToFlags(kafkarestv3.ACLRESOURCETYPE_TOPIC, kafkarestv3.ACLRESOURCETYPE_GROUP,
				kafkarestv3.ACLRESOURCETYPE_CLUSTER, kafkarestv3.ACLRESOURCETYPE_TRANSACTIONAL_ID)))
	}
	return aclConfiguration
}

func aclFlags() *pflag.FlagSet {
	flgSet := pflag.NewFlagSet("acl-config", pflag.ExitOnError)
	flgSet.String("kafka-cluster-id", "", "Kafka cluster ID for scope of ACL commands.")
	flgSet.Bool("allow", false, "ACL permission to allow access.")
	flgSet.Bool("deny", false, "ACL permission to restrict access to resource.")
	flgSet.String("principal", "", "Principal for this operation with User: or Group: prefix.")
	flgSet.String("host", "*", "Set host for access. Only IP addresses are supported.")
	flgSet.String("operation", "", fmt.Sprintf("Set ACL Operation to: (%s).",
		convertToFlags(kafkarestv3.ACLOPERATION_ALL, kafkarestv3.ACLOPERATION_READ, kafkarestv3.ACLOPERATION_WRITE,		// TODO : do we want the ALL operation included? (included w iam acl but not in cloud)
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
	flgSet.SortFlags = false
	return flgSet
}

func convertToFlags(operations ...interface{}) string {
	var ops []string

	for _, v := range operations {
		if v == mds.ACLRESOURCETYPE_GROUP {
			v = "consumer-group"
		}
		if v == mds.ACLRESOURCETYPE_CLUSTER {
			v = "cluster-scope"
		}
		s := fmt.Sprintf("%v", v)
		s = strings.ReplaceAll(s, "_", "-")
		ops = append(ops, strings.ToLower(s))
	}

	sort.Strings(ops)
	return strings.Join(ops, ", ")
}

func parseCreatAclRequest(cmd *cobra.Command) *CreateAclRequestDataWithError {
	aclRequest := &CreateAclRequestDataWithError{
		CreateAclRequestData: &kafkarestv3.CreateAclRequestData{
			Host: "*",
		},
		errors: nil,
	}
	cmd.Flags().Visit(populateCreateAclRequest(aclRequest))
	return aclRequest
}

func populateCreateAclRequest(conf *CreateAclRequestDataWithError) func(*pflag.Flag) {
	return func(flag *pflag.Flag) {
		v := flag.Value.String()
		switch n := flag.Name; n {
		case "consumer-group":
			setCreateAclRequestResourcePattern(conf, "GROUP", v) // set aclRequestData.ResourceType and aclRequestData.ResourceName
		case "cluster-scope":
			// The only valid name for a cluster is kafka-cluster
			// https://github.com/confluentinc/cc-kafka/blob/88823c6016ea2e306340938994d9e122abf3c6c0/core/src/main/scala/kafka/security/auth/Resource.scala#L24
			setCreateAclRequestResourcePattern(conf, "cluster", "kafka-cluster")
		case "topic":
			fallthrough
		case "delegation-token":
			fallthrough
		case "transactional-id":
			setCreateAclRequestResourcePattern(conf, n, v)
		case "allow":
			conf.Permission = kafkarestv3.ACLPERMISSION_ALLOW
		case "deny":
			conf.Permission = kafkarestv3.ACLPERMISSION_DENY
		case "prefix":
			conf.PatternType = kafkarestv3.ACLPATTERNTYPE_PREFIXED
		case "principal":
			conf.Principal = v
		case "host":
			conf .Host = v
		}
	}
}

func setCreateAclRequestResourcePattern(conf *CreateAclRequestDataWithError, n, v string) {
	if conf.ResourceType != "" {
		// A resourceType has already been set with a previous flag
		conf.errors = multierror.Append(conf.errors, fmt.Errorf("exactly one of %v must be set",
			convertToFlags(kafkarestv3.ACLRESOURCETYPE_TOPIC, kafkarestv3.ACLRESOURCETYPE_GROUP,
				kafkarestv3.ACLRESOURCETYPE_CLUSTER, kafkarestv3.ACLRESOURCETYPE_TRANSACTIONAL_ID)))
		return
	}

	// Normalize the resource pattern name
	n = strings.ToUpper(n)
	n = strings.ReplaceAll(n, "-", "_")

	//enumUtils := enumUtils{}
	//enumUtils.init(mds.ACLRESOURCETYPE_TOPIC, mds.ACLRESOURCETYPE_GROUP,
	//	mds.ACLRESOURCETYPE_CLUSTER, mds.ACLRESOURCETYPE_TRANSACTIONAL_ID)
	//conf.AclBinding.Pattern.ResourceType = enumUtils[n].(mds.AclResourceType)
	conf.ResourceType = kafkarestv3.AclResourceType(n) // TODO does this work?

	if conf.ResourceType == kafkarestv3.ACLRESOURCETYPE_CLUSTER {
		conf.PatternType = kafkarestv3.ACLPATTERNTYPE_LITERAL
	}
	conf.ResourceName = v
}
