package kafka

import (
	"context"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

var quotaListFields = []string{"DisplayName", "Description", "Throughput", "Cluster", "Principals"}

func (c *quotaCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cloud provider regions.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *quotaCommand) list(cmd *cobra.Command, _ []string) error {

	// TODO figure out how to specify Service Account
	//sa, err := cmd.Flags().GetString("service-account")
	//if err != nil {
	//	return err
	//}

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	//fmt.Println(sa)
	//fmt.Println(cluster)

	req := c.V2Client.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.ListKafkaQuotasV1ClientQuotas(context.Background())
	kafkaQuotaList, _, err := req.Cluster(cluster.ID).Execute()
	if err != nil {
		return err
	}

	w, err := output.NewListOutputWriter(cmd, quotaListFields, camelToSpaced(quotaListFields), camelToSnake(quotaListFields))
	if err != nil {
		return err
	}

	for _, quota := range kafkaQuotaList.Data {
		w.AddElement(quota)
	}

	return w.Out()
}
