package iam

import (
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *aclCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Kafka ACL.",
		Long:  "Create a Kafka ACL. This command only works with centralized ACLs.",
		Args:  cobra.NoArgs,
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create an ACL that grants the specified user "read" permission to the specified consumer group in the specified Kafka cluster:`,
				Code: "confluent iam acl create --allow --principal User:User1 --operation read --consumer-group java_example_group_1 --kafka-cluster <kafka-cluster-id>",
			},
			examples.Example{
				Text: `Create an ACL that grants the specified user "write" permission on all topics in the specified Kafka cluster:`,
				Code: "confluent iam acl create --allow --principal User:User1 --operation write --topic '*' --kafka-cluster <kafka-cluster-id>",
			},
			examples.Example{
				Text: `Create an ACL that assigns a group "read" access to all topics that use the specified prefix in the specified Kafka cluster:`,
				Code: "confluent iam acl create --allow --principal Group:Finance --operation read --topic financial --prefix --kafka-cluster <kafka-cluster-id>",
			},
		),
	}

	cmd.Flags().AddFlagSet(aclFlags())
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("kafka-cluster"))
	cobra.CheckErr(cmd.MarkFlagRequired("principal"))
	cobra.CheckErr(cmd.MarkFlagRequired("operation"))

	return cmd
}

func (c *aclCommand) create(cmd *cobra.Command, _ []string) error {
	acl := validateACLAddDelete(parse(cmd))
	if acl.errors != nil {
		return acl.errors
	}

	response, err := c.MDSClient.KafkaACLManagementApi.AddAclBinding(c.createContext(), *acl.CreateAclRequest)
	if err != nil {
		return c.handleACLError(cmd, err, response)
	}

	return printACLs(cmd, acl.CreateAclRequest.Scope.Clusters.KafkaCluster, []mds.AclBinding{acl.CreateAclRequest.AclBinding})
}

func validateACLAddDelete(aclConfiguration *ACLConfiguration) *ACLConfiguration {
	// delete is deliberately less powerful in the cli than in the API to prevent accidental
	// deletion of too many acls at once. Expectation is that multi delete will be done via
	// repeated invocation of the cli by external scripts.
	if aclConfiguration.AclBinding.Entry.PermissionType == "" {
		aclConfiguration.errors = multierror.Append(aclConfiguration.errors, errors.Errorf(errors.MustSetAllowOrDenyErrorMsg))
	}

	if aclConfiguration.AclBinding.Pattern.PatternType == "" {
		aclConfiguration.AclBinding.Pattern.PatternType = mds.PATTERNTYPE_LITERAL
	}

	if aclConfiguration.AclBinding.Pattern.ResourceType == "" {
		aclConfiguration.errors = multierror.Append(aclConfiguration.errors, errors.Errorf(errors.MustSetResourceTypeErrorMsg,
			convertToFlags(mds.ACLRESOURCETYPE_TOPIC, mds.ACLRESOURCETYPE_GROUP,
				mds.ACLRESOURCETYPE_CLUSTER, mds.ACLRESOURCETYPE_TRANSACTIONAL_ID)))
	}
	return aclConfiguration
}
