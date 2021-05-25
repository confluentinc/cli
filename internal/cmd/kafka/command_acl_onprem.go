package kafka

import (
	"fmt"
	aclutil "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	onPremAclListFields = []string{"Principal", "Permission", "Operation", "Host", "ResourceType", "ResourceName", "PatternType"}
	onPremAclListStructuredRenames = []string{"principal", "permission", "operation", "host", "resource_type", "resource_name", "pattern_type"}
)

type aclOnPremCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func NewAclCommandOnPrem(prerunner pcmd.PreRunner) *cobra.Command {
	aclCmd := &aclOnPremCommand{
		pcmd.NewAuthenticatedStateFlagCommand(
			&cobra.Command{
				Use: "acl",
				Short: "Manage Kafka ACLs.",
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
				Text: "You can specify only one of the following flags per command invocation: ``cluster-scope``, ``consumer-group``, ``topic``, or ``transactional-id``. For example, to modify both ``consumer-group`` and ``topic`` resources, you need to issue two separate commands:",
				Code: "confluent kafka acl create --allow --User:Jane --operation READ --consumer-group java_example_group_1\nconfluent kafka acl create --allow --Group:Finance --operation READ --topic '*'",
			}),
	}
	createCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	createCmd.Flags().AddFlagSet(createACLFlags())
	createCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	createCmd.Flags().SortFlags = false
	aclCmd.AddCommand(createCmd)

	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete Kafka ACLs matching the search criteria.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(aclCmd.delete),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete all READ access ACLs for the specified user:",
				Code: "confluent kafka acl delete --operation READ --allow --topic Test --principal User:Jane --host '*'",
			}),
	}
	deleteCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	deleteCmd.Flags().AddFlagSet(deleteACLFlags())
	deleteCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	deleteCmd.Flags().SortFlags = false
	aclCmd.AddCommand(deleteCmd)

	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs.",
		Args: cobra.NoArgs,
		RunE: pcmd.NewCLIRunE(aclCmd.list),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all the local ACLs for the Kafka cluster:",
				Code: "confluent acl list",
			},
			examples.Example{
				Text: "List all the ACLs for the Kafka cluster that include allow permissions for the user Jane:",
				Code: "confluent kafka acl list --allow --principal User:Jane",
			},
		),
	}
	listCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	listCmd.Flags().AddFlagSet(aclFlags())
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	aclCmd.AddCommand(listCmd)
}

func (aclCmd *aclOnPremCommand) list(cmd *cobra.Command, _ []string) error {
	acl := aclutil.ParseAclRequest(cmd)
	if acl.Errors != nil {
		return acl.Errors
	}
	restClient, restContext, err := initKafkaRest(aclCmd.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	opts := aclutil.AclRequestToListAclReqest(acl)
	aclGetResp, httpResp, err := restClient.ACLApi.ClustersClusterIdAclsGet(restContext, clusterId, opts)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, httpResp)
	}
	return aclutil.PrintACLsFromKafkaRestResponse(cmd, aclGetResp.Data, cmd.OutOrStdout(), onPremAclListFields, onPremAclListStructuredRenames)
}

func (aclCmd *aclOnPremCommand) create(cmd *cobra.Command, _ []string) error {
	acl := aclutil.ParseAclRequest(cmd)
	acl = validateCreateDeleteAclRequestData(acl)
	if acl.Errors != nil {
		return acl.Errors
	}
	restClient, restContext, err := initKafkaRest(aclCmd.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	opts := aclutil.AclRequestToCreateAclReqest(acl)
	httpResp, err := restClient.ACLApi.ClustersClusterIdAclsPost(restContext, clusterId, opts)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, httpResp)
	}
	aclData := aclutil.CreateAclRequestDataToAclData(acl)
	return aclutil.PrintACLsFromKafkaRestResponse(cmd, []kafkarestv3.AclData{aclData}, cmd.OutOrStdout(), onPremAclListFields, onPremAclListStructuredRenames)
}

func (aclCmd *aclOnPremCommand) delete(cmd *cobra.Command, _ []string) error {
	acl := aclutil.ParseAclRequest(cmd)
	acl = validateCreateDeleteAclRequestData(acl)
	if acl.Errors != nil {
		return acl.Errors
	}
	restClient, restContext, err := initKafkaRest(aclCmd.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	opts := aclutil.AclRequestToDeleteAclReqest(acl)
	aclDeleteResp, httpResp, err := restClient.ACLApi.ClustersClusterIdAclsDelete(restContext, clusterId, opts)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, httpResp)
	}
	return aclutil.PrintACLsFromKafkaRestResponse(cmd, aclDeleteResp.Data, cmd.OutOrStdout(), onPremAclListFields, onPremAclListStructuredRenames)
}

func validateCreateDeleteAclRequestData(aclConfiguration *aclutil.AclRequestDataWithError) *aclutil.AclRequestDataWithError {
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
			aclutil.ConvertToFlags(kafkarestv3.ACLRESOURCETYPE_TOPIC, kafkarestv3.ACLRESOURCETYPE_GROUP,
				kafkarestv3.ACLRESOURCETYPE_CLUSTER, kafkarestv3.ACLRESOURCETYPE_TRANSACTIONAL_ID)))
	}
	return aclConfiguration
}

func createACLFlags() *pflag.FlagSet {
	flgSet := aclFlags()
	_ = cobra.MarkFlagRequired(flgSet, "principal")
	_ = cobra.MarkFlagRequired(flgSet, "operation")
	return flgSet
}

func deleteACLFlags() *pflag.FlagSet {
	flgSet := aclFlags()
	_ = cobra.MarkFlagRequired(flgSet, "principal")
	_ = cobra.MarkFlagRequired(flgSet, "operation")
	_ = cobra.MarkFlagRequired(flgSet, "host")
	return flgSet
}

func aclFlags() *pflag.FlagSet {
	flgSet := pflag.NewFlagSet("acl-config", pflag.ExitOnError)
	flgSet.Bool("allow", false, "ACL permission to allow access.")
	flgSet.Bool("deny", false, "ACL permission to restrict access to resource.")
	flgSet.String("principal", "", "Principal for this operation with User: or Group: prefix.")
	flgSet.String("host", "*", "Set host for access. Only IP addresses are supported.")
	flgSet.String("operation", "", fmt.Sprintf("Set ACL Operation to: (%s).",
		aclutil.ConvertToFlags(kafkarestv3.ACLOPERATION_ALL, kafkarestv3.ACLOPERATION_READ, kafkarestv3.ACLOPERATION_WRITE,		// TODO : do we want the ALL operation included? (included w iam acl but not in cloud)
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
