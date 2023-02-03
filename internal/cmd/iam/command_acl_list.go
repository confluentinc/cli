package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *aclCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs for a resource.",
		Long:  "List Kafka ACLs for a resource. This command only works with centralized ACLs.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all the ACLs for the specified Kafka cluster:",
				Code: "confluent iam acl list --kafka-cluster <kafka-cluster-id>",
			},
			examples.Example{
				Text: `List all the ACLs for the specified cluster that include "allow" permissions for the user Jane:`,
				Code: "confluent iam acl list --kafka-cluster <kafka-cluster-id> --allow --principal User:Jane",
			},
		),
	}

	cmd.Flags().AddFlagSet(aclFlags())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("kafka-cluster")

	return cmd
}

func (c *aclCommand) list(cmd *cobra.Command, _ []string) error {
	acl := parse(cmd)

	bindings, response, err := c.MDSClient.KafkaACLManagementApi.SearchAclBinding(c.createContext(), convertToACLFilterRequest(acl.CreateAclRequest))
	if err != nil {
		return c.handleACLError(cmd, err, response)
	}

	return printACLs(cmd, acl.Scope.Clusters.KafkaCluster, bindings)
}
