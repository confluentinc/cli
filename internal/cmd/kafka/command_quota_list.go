package kafka

import (
	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/kafka-quotas/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	quotaListFields       = []string{"Id", "DisplayName", "Description", "Ingress", "Egress", "Cluster", "Principals"}
	quotaHumanFields      = []string{"ID", "Display Name", "Description", "Ingress", "Egress", "Cluster", "Principals"}
	quotaStructuredFields = []string{"id", "display_name", "description", "ingress", "egress", "cluster", "principals"}
)

func (c *quotaCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List client quotas for given cluster.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(examples.Example{
			Text: `List client quotas for cluster "lkc-12345".`,
			Code: `confluent kafka quota list --cluster lkc-12345`,
		}),
	}

	pcmd.AddPrincipalFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *quotaCommand) list(cmd *cobra.Command, _ []string) error {
	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	quotas, err := c.V2Client.ListKafkaQuotas(cluster.ID, c.EnvironmentId())
	if err != nil {
		return quotaErr(err)
	}

	w, err := output.NewListOutputWriter(cmd, quotaListFields, quotaHumanFields, quotaStructuredFields)
	if err != nil {
		return err
	}

	// TODO use API for filtering by principal when it becomes available: https://confluentinc.atlassian.net/browse/KPLATFORM-733
	if cmd.Flags().Changed("principal") {
		principal, err := cmd.Flags().GetString("principal")
		if err != nil {
			return err
		}
		quotas = getQuotasForPrincipal(quotas, principal)
	}
	format, _ := cmd.Flags().GetString(output.FlagName)
	for _, quota := range quotas {
		w.AddElement(quotaToPrintable(quota, format))
	}

	return w.Out()
}

func getQuotasForPrincipal(quotas []v1.KafkaQuotasV1ClientQuota, principal string) []v1.KafkaQuotasV1ClientQuota {
	var filteredQuotaData []v1.KafkaQuotasV1ClientQuota
	for _, quota := range quotas {
		for _, p := range *quota.Principals {
			if p.Id == principal {
				filteredQuotaData = append(filteredQuotaData, quota)
				// principals can only belong to one quota so break after finding it
				return filteredQuotaData
			}
		}
	}
	return filteredQuotaData
}
