package kafka

import (
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	aclutil "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

var (
	onPremAclListFields            = []string{"Principal", "Permission", "Operation", "Host", "ResourceType", "ResourceName", "PatternType"}
	onPremAclListStructuredRenames = []string{"principal", "permission", "operation", "host", "resource_type", "resource_name", "pattern_type"}
)

func (c *aclCommand) onPremInit() {
	createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a Kafka ACL.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.onPremCreate),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "You can specify only one of the following flags per command invocation: `cluster-scope`, `consumer-group`, `topic`, or `transactional-id`. For example, for a consumer to read a topic, you need to grant `READ` and `DESCRIBE` both on the `consumer-group` and the `topic` resources, issuing two separate commands:",
				Code: "confluent kafka acl create --allow --principal User:Jane --operation READ --operation DESCRIBE --consumer-group java_example_group_1\nconfluent kafka acl create --allow --Group:Finance --operation READ --operation DESCRIBE --topic '*'",
			}),
	}
	createCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	createCmd.Flags().AddFlagSet(aclutil.CreateACLFlags())
	pcmd.AddOutputFlag(createCmd)
	c.AddCommand(createCmd)

	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete Kafka ACLs matching the search criteria.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.onPremDelete),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete all READ access ACLs for the specified user:",
				Code: "confluent kafka acl delete --operation READ --allow --topic Test --principal User:Jane --host '*'",
			}),
	}
	deleteCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	deleteCmd.Flags().AddFlagSet(aclutil.DeleteACLFlags())
	pcmd.AddOutputFlag(deleteCmd)
	c.AddCommand(deleteCmd)

	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.onPremList),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all the local ACLs for the Kafka cluster:",
				Code: "confluent kafka acl list",
			},
			examples.Example{
				Text: "List all the ACLs for the Kafka cluster that include allow permissions for the user Jane:",
				Code: "confluent kafka acl list --allow --principal User:Jane",
			},
		),
	}
	listCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	listCmd.Flags().AddFlagSet(aclutil.AclFlags())
	pcmd.AddOutputFlag(listCmd)
	c.AddCommand(listCmd)
}

func (c *aclCommand) onPremList(cmd *cobra.Command, _ []string) error {
	acl := aclutil.ParseAclRequest(cmd)
	if acl.Errors != nil {
		return acl.Errors
	}
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
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

func (c *aclCommand) onPremCreate(cmd *cobra.Command, _ []string) error {
	acl := aclutil.ParseAclRequest(cmd)
	acl = aclutil.ValidateCreateDeleteAclRequestData(acl)
	if acl.Errors != nil {
		return acl.Errors
	}
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
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

func (c *aclCommand) onPremDelete(cmd *cobra.Command, _ []string) error {
	acl := aclutil.ParseAclRequest(cmd)
	acl = aclutil.ValidateCreateDeleteAclRequestData(acl)
	if acl.Errors != nil {
		return acl.Errors
	}
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
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
