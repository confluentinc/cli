package kafka

import (
	"github.com/spf13/cobra"

	kafkaquotasv1 "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
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
	quotas, err := c.getQuotas()
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("principal") {
		principal, err := cmd.Flags().GetString("principal")
		if err != nil {
			return err
		}
		quotas = filterQuotasByPrincipal(quotas, principal)
	}

	list := output.NewList(cmd)
	format := output.GetFormat(cmd)
	for _, quota := range quotas {
		list.Add(quotaToPrintable(quota, format))
	}
	return list.Print()
}

func filterQuotasByPrincipal(quotas []kafkaquotasv1.KafkaQuotasV1ClientQuota, principalId string) []kafkaquotasv1.KafkaQuotasV1ClientQuota {
	var filteredQuotas []kafkaquotasv1.KafkaQuotasV1ClientQuota
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
