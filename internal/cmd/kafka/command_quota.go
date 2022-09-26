package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	launchdarkly "github.com/confluentinc/cli/internal/pkg/featureflags"
	"github.com/confluentinc/cli/internal/pkg/output"

	kafkaquotas "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"
)

type quotaCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func newQuotaCommand(config *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "quota",
		Short:       "Manage Kafka client quotas.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &quotaCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}

	dc := dynamicconfig.New(config, nil, nil)
	_ = dc.ParseFlagsIntoConfig(cmd)

	c.Hidden = !(config.IsTest || launchdarkly.Manager.BoolVariation("cli.client_quotas.enable", dc.Context(), v1.CliLaunchDarklyClient, true, false))

	c.AddCommand(c.newCreateCommand())
	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newUpdateCommand())

	return c.Command
}

type quotaOut struct {
	Id          string `human:"ID" serialized:"id"row`
	DisplayName string `human:"Name" serialized:"display_name"row`
	Description string `human:"Description" serialized:"description"row`
	Ingress     string `human:"Ingress" serialized:"ingress"row`
	Egress      string `human:"Egress" serialized:"egress"row`
	Principals  string `human:"Principals" serialized:"principals"row`
	Cluster     string `human:"Cluster" serialized:"cluster"row`
	Environment string `human:"Environment" serialized:"environment"row`
}

func quotaToPrintable(quota kafkaquotas.KafkaQuotasV1ClientQuota, format output.Format) *quotaOut {
	s := &quotaOut{
		Id:          quota.GetId(),
		DisplayName: quota.GetDisplayName(),
		Description: quota.GetDescription(),
		Ingress:     quota.Throughput.GetIngressByteRate(),
		Egress:      quota.Throughput.GetEgressByteRate(),
		Principals:  principalsToString(quota.GetPrincipals()),
		Cluster:     quota.Cluster.Id,
		Environment: quota.Environment.Id,
	}
	if format == output.Human {
		s.Ingress += " B/s"
		s.Egress += " B/s"
	}
	return s
}

func principalsToString(principals []kafkaquotas.ObjectReference) string {
	principalStr := ""
	for i, principal := range principals {
		principalStr += principal.Id
		if i < len(principals)-1 {
			principalStr += ","
		}
	}
	return principalStr
}
