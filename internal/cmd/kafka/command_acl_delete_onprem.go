package kafka

import (
	"github.com/spf13/cobra"

	aclutil "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *aclCommand) newDeleteCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete Kafka ACLs matching the search criteria.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.deleteOnPrem),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete all READ access ACLs for the specified user:",
				Code: "confluent kafka acl delete --operation READ --allow --topic Test --principal User:Jane --host '*'",
			},
		),
	}

	cmd.Flags().AddFlagSet(aclutil.AclFlags())
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("principal")
	_ = cmd.MarkFlagRequired("operation")
	_ = cmd.MarkFlagRequired("host")

	return cmd
}

func (c *aclCommand) deleteOnPrem(cmd *cobra.Command, _ []string) error {
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
	aclDeleteResp, httpResp, err := restClient.ACLV3Api.DeleteKafkaAcls(restContext, clusterId, opts)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, httpResp)
	}

	return aclutil.PrintACLsFromKafkaRestResponse(cmd, aclDeleteResp.Data, cmd.OutOrStdout(), listFieldsOnPrem, humanLabelsOnPrem, structuredLabelsOnPrem)
}
