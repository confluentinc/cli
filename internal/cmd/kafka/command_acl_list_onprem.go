package kafka

import (
	"github.com/spf13/cobra"

	aclutil "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *aclCommand) newListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.listOnPrem),
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

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	cmd.Flags().AddFlagSet(aclutil.AclFlags())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *aclCommand) listOnPrem(cmd *cobra.Command, _ []string) error {
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

	return aclutil.PrintACLsFromKafkaRestResponse(cmd, aclGetResp.Data, cmd.OutOrStdout(), listFieldsOnPrem, listStructuredRenamesOnPrem)
}
