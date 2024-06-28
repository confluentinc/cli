package kafka

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	kafkaquotasv1 "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/kafka"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type quotaCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newQuotaCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "quota",
		Short:       "Manage Kafka client quotas.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &quotaCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

type quotaHumanOut struct {
	Id          string `human:"ID"`
	DisplayName string `human:"Name"`
	Description string `human:"Description"`
	Ingress     string `human:"Ingress (B/s)"`
	Egress      string `human:"Egress (B/s)"`
	Principals  string `human:"Principals"`
	Cluster     string `human:"Cluster"`
	Environment string `human:"Environment"`
}

type quotaSerializedOut struct {
	Id          string   `serialized:"id"`
	DisplayName string   `serialized:"name"`
	Description string   `serialized:"description"`
	Ingress     string   `serialized:"ingress"`
	Egress      string   `serialized:"egress"`
	Principals  []string `serialized:"principals"`
	Cluster     string   `serialized:"cluster"`
	Environment string   `serialized:"environment"`
}

func printTable(cmd *cobra.Command, quota kafkaquotasv1.KafkaQuotasV1ClientQuota) error {
	table := output.NewTable(cmd)
	if output.GetFormat(cmd) == output.Human {
		table.Add(&quotaHumanOut{
			Id:          quota.GetId(),
			DisplayName: quota.Spec.GetDisplayName(),
			Description: quota.Spec.GetDescription(),
			Ingress:     quota.Spec.Throughput.GetIngressByteRate(),
			Egress:      quota.Spec.Throughput.GetEgressByteRate(),
			Principals:  strings.Join(principalsToStringSlice(quota.Spec.GetPrincipals()), ", "),
			Cluster:     quota.Spec.Cluster.GetId(),
			Environment: quota.Spec.Environment.GetId(),
		})
	} else {
		table.Add(&quotaSerializedOut{
			Id:          quota.GetId(),
			DisplayName: quota.Spec.GetDisplayName(),
			Description: quota.Spec.GetDescription(),
			Ingress:     quota.Spec.Throughput.GetIngressByteRate(),
			Egress:      quota.Spec.Throughput.GetEgressByteRate(),
			Principals:  principalsToStringSlice(quota.Spec.GetPrincipals()),
			Cluster:     quota.Spec.Cluster.GetId(),
			Environment: quota.Spec.Environment.GetId(),
		})
	}

	return table.Print()
}

func principalsToStringSlice(principals []kafkaquotasv1.GlobalObjectReference) []string {
	ids := make([]string, len(principals))
	for i, principal := range principals {
		principalIds[i] = principal.GetId()
	}

	return principalIds
}

func (c *quotaCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validArgsMultiple(cmd, args)
}

func (c *quotaCommand) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteQuotas()
}

func (c *quotaCommand) autocompleteQuotas() []string {
	quotas, err := c.getQuotas()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(quotas))
	for i, quota := range quotas {
		description := fmt.Sprintf("%s: %s", quota.Spec.GetDisplayName(), quota.Spec.GetDescription())
		suggestions[i] = fmt.Sprintf("%s\t%s", quota.GetId(), description)
	}
	return suggestions
}

func (c *quotaCommand) getQuotas() ([]kafkaquotasv1.KafkaQuotasV1ClientQuota, error) {
	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return nil, err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListKafkaQuotas(cluster.ID, environmentId)
}
