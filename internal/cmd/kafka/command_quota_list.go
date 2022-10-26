package kafka

import (
	v1 "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	quotaListFields       = []string{"Id", "DisplayName", "Description", "Ingress", "Egress", "Cluster", "Principals", "Environment"}
	quotaHumanFields      = []string{"ID", "Name", "Description", "Ingress", "Egress", "Cluster", "Principals", "Environment"}
	quotaStructuredFields = []string{"id", "name", "description", "ingress", "egress", "cluster", "principals", "environment"}
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
		return err
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
		quotas = filterQuotasByPrincipal(quotas, principal)
	}
	format, _ := cmd.Flags().GetString(output.FlagName)
	for _, quota := range quotas {
		w.AddElement(quotaToPrintable(quota, format))
	}

	return w.Out()
}

func filterQuotasByPrincipal(quotas []v1.KafkaQuotasV1ClientQuota, principalId string) []v1.KafkaQuotasV1ClientQuota {
	var filteredQuotas []v1.KafkaQuotasV1ClientQuota
	for _, quota := range quotas {
		for _, principal := range *quota.Spec.Principals {
			if principal.Id == principalId {
				filteredQuotas = append(filteredQuotas, quota)
				// principals can only belong to one quota so break after finding it
				return filteredQuotas
			}
		}
	}
	return filteredQuotas
}
