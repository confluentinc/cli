package kafka

import (
	"github.com/spf13/cobra"

	aclutil "github.com/confluentinc/cli/v4/pkg/acl"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *aclCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs for a resource.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	cmd.Flags().Bool("cluster-scope", false, "Modify ACLs for the cluster.")
	cmd.Flags().String("topic", "", "Modify ACLs for the specified topic resource.")
	cmd.Flags().String("consumer-group", "", "Modify ACLs for the specified consumer group resource.")
	cmd.Flags().String("transactional-id", "", "Modify ACLs for the specified TransactionalID resource.")
	cmd.Flags().Bool("prefix", false, "When this flag is set, the specified resource name is interpreted as a prefix.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("principal", "", `Principal for this operation, prefixed with "User:".`)
	cmd.Flags().Bool("all", false, "Include ACLs for deleted principals with integer IDs.")
	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)
	cmd.MarkFlagsMutuallyExclusive("service-account", "principal")

	return cmd
}

func (c *aclCommand) list(cmd *cobra.Command, _ []string) error {
	acl, err := parse(c.Context, cmd)
	if err != nil {
		return err
	}

	if acl[0].errors != nil {
		return acl[0].errors
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	if err := c.provisioningClusterCheck(kafkaREST.GetClusterId()); err != nil {
		return err
	}

	aclDataList, err := kafkaREST.CloudClient.GetKafkaAcls(acl[0].ACLBinding)
	if err != nil {
		return err
	}

	return aclutil.PrintACLsFromKafkaRestResponse(cmd, aclDataList.Data)
}
