package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *aclCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List centralized ACLs for a resource.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all the ACLs for the specified Kafka cluster:",
				Code: "confluent iam acl list --kafka-cluster <kafka-cluster-id>",
			},
			examples.Example{
				Text: `List all the ACLs for the specified cluster that include "allow" permissions for user "Jane":`,
				Code: "confluent iam acl list --kafka-cluster <kafka-cluster-id> --allow --principal User:Jane",
			},
		),
	}

	cmd.Flags().AddFlagSet(aclFlags())
	pcmd.AddMDSOnPremMTLSFlags(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("kafka-cluster"))

	return cmd
}

func (c *aclCommand) list(cmd *cobra.Command, _ []string) error {
	client, err := c.GetMDSClient(cmd)
	if err != nil {
		return err
	}

	acl := parse(cmd)

	bindings, response, err := client.KafkaACLManagementApi.SearchAclBinding(c.createContext(), convertToAclFilterRequest(acl.CreateAclRequest))
	if err != nil {
		return c.handleAclError(cmd, err, response)
	}

	return printAcls(cmd, acl.Scope.Clusters.KafkaCluster, bindings)
}
