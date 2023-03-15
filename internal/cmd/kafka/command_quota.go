package kafka

import (
	"github.com/spf13/cobra"

	kafkaquotas "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type quotaCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func newQuotaCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "quota",
		Short:       "Manage Kafka client quotas.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &quotaCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}

	c.AddCommand(c.newCreateCommand())
	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newUpdateCommand())

	return c.Command
}

type quotaOut struct {
	Id          string `human:"ID" serialized:"id"`
	DisplayName string `human:"Name" serialized:"name"`
	Description string `human:"Description" serialized:"description"`
	Ingress     string `human:"Ingress" serialized:"ingress"`
	Egress      string `human:"Egress" serialized:"egress"`
	Principals  string `human:"Principals" serialized:"principals"`
	Cluster     string `human:"Cluster" serialized:"cluster"`
	Environment string `human:"Environment" serialized:"environment"`
}

func quotaToPrintable(quota kafkaquotas.KafkaQuotasV1ClientQuota, format output.Format) *quotaOut {
	out := &quotaOut{
		Id:          quota.GetId(),
		DisplayName: quota.Spec.GetDisplayName(),
		Description: quota.Spec.GetDescription(),
		Ingress:     quota.Spec.Throughput.GetIngressByteRate(),
		Egress:      quota.Spec.Throughput.GetEgressByteRate(),
		Principals:  principalsToString(quota.Spec.GetPrincipals()),
		Cluster:     quota.Spec.Cluster.GetId(),
		Environment: quota.Spec.Environment.GetId(),
	}
	if format == output.Human {
		out.Ingress += " B/s"
		out.Egress += " B/s"
	}
	return out
}

func principalsToString(principals []kafkaquotas.GlobalObjectReference) string {
	principalStr := ""
	for i, principal := range principals {
		principalStr += principal.Id
		if i < len(principals)-1 {
			principalStr += ","
		}
	}
	return principalStr
}
