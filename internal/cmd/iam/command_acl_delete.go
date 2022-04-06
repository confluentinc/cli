package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *aclCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Kafka ACL.",
		Long:  "Delete a Kafka ACL. This command only works with centralized ACLs.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.delete),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete an ACL that granted the specified user access to the "test" topic in the specified cluster.`,
				Code: "confluent iam acl delete --kafka-cluster-id <kafka-cluster-id> --allow --principal User:Jane --topic test --operation write --host *",
			},
		),
	}

	cmd.Flags().AddFlagSet(aclFlags())
	pcmd.AddContextFlag(cmd, c.CLICommand)

	_ = cmd.MarkFlagRequired("kafka-cluster-id")
	_ = cmd.MarkFlagRequired("principal")
	_ = cmd.MarkFlagRequired("operation")
	_ = cmd.MarkFlagRequired("host")

	return cmd
}

func (c *aclCommand) delete(cmd *cobra.Command, _ []string) error {
	acl := parse(cmd)
	if acl.errors != nil {
		return acl.errors
	}

	bindings, response, err := c.MDSClient.KafkaACLManagementApi.RemoveAclBindings(c.createContext(), convertToACLFilterRequest(acl.CreateAclRequest))
	if err != nil {
		return c.handleACLError(cmd, err, response)
	}

	return printACLs(cmd, acl.Scope.Clusters.KafkaCluster, bindings)
}
