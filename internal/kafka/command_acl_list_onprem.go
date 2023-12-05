package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/acl"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
)

func (c *aclCommand) newListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs.",
		Args:  cobra.NoArgs,
		RunE:  c.listOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all the local ACLs for the Kafka cluster (providing embedded Kafka REST Proxy endpoint).",
				Code: "confluent kafka acl list --url http://localhost:8090/kafka",
			},
			examples.Example{
				Text: "List all the local ACLs for the Kafka cluster (providing Kafka REST Proxy endpoint).",
				Code: "confluent kafka acl list --url http://localhost:8082",
			},
			examples.Example{
				Text: `List all the ACLs for the Kafka cluster that include allow permissions for the user "Jane":`,
				Code: "confluent kafka acl list --allow --principal User:Jane",
			},
		),
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	cmd.Flags().AddFlagSet(acl.Flags())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *aclCommand) listOnPrem(cmd *cobra.Command, _ []string) error {
	data := acl.ParseRequest(cmd)
	if data.Errors != nil {
		return data.Errors
	}

	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	opts := acl.RequestToListRequest(data)
	aclGetResp, httpResp, err := restClient.ACLV3Api.GetKafkaAcls(restContext, clusterId, opts)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, httpResp)
	}

	return acl.PrintACLsFromKafkaRestResponseOnPrem(cmd, aclGetResp.Data)
}
