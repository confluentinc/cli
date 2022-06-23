package kafka

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

var (
	quotaListFields       = []string{"Id", "DisplayName", "Description", "Ingress", "Egress", "Cluster", "Principals"}
	quotaHumanFields      = []string{"ID", "Display Name", "Description", "Ingress", "Egress", "Cluster", "Principals"}
	quotaStructuredFields = []string{"id", "display_name", "description", "ingress", "egress", "cluster", "principals"}
)

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
	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	// TODO pagination

	req := c.V2Client.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.ListKafkaQuotasV1ClientQuotas(c.quotaContext())
	kafkaQuotaList, _, err := req.Cluster(cluster.ID).Environment(c.EnvironmentId()).Execute()
	if err != nil {
		return quotaErr(err)
	}

	w, err := output.NewListOutputWriter(cmd, quotaListFields, quotaHumanFields, quotaStructuredFields)
	if err != nil {
		return err
	}

	for _, quota := range kafkaQuotaList.Data {
		w.AddElement(quotaToPrintable(quota))
	}

	return w.Out()
}
