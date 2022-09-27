package kafka

import (
	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
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

	// TODO use API for filtering by principal when it becomes available: https://confluentinc.atlassian.net/browse/KPLATFORM-733
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

func filterQuotasByPrincipal(quotas []v1.KafkaQuotasV1ClientQuota, principal string) []v1.KafkaQuotasV1ClientQuota {
	var filteredQuotas []v1.KafkaQuotasV1ClientQuota
	for _, quota := range quotas {
		for _, p := range *quota.Principals {
			if p.Id == principal {
				filteredQuotas = append(filteredQuotas, quota)
				// principals can only belong to one quota so break after finding it
				return filteredQuotas
			}
		}
	}
	return filteredQuotas
}
