package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	kafkaquotasv1 "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/output"
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

type quotaOut struct {
	Id          string   `human:"ID" serialized:"id"`
	DisplayName string   `human:"Name" serialized:"name"`
	Description string   `human:"Description" serialized:"description"`
	Ingress     string   `human:"Ingress (B/s)" serialized:"ingress"`
	Egress      string   `human:"Egress (B/s)" serialized:"egress"`
	Principals  []string `human:"Principals" serialized:"principals"`
	Cluster     string   `human:"Cluster" serialized:"cluster"`
	Environment string   `human:"Environment" serialized:"environment"`
}

func printTable(cmd *cobra.Command, quota kafkaquotasv1.KafkaQuotasV1ClientQuota) error {
	table := output.NewTable(cmd)
	table.Add(&quotaOut{
		Id:          quota.GetId(),
		DisplayName: quota.Spec.GetDisplayName(),
		Description: quota.Spec.GetDescription(),
		Ingress:     quota.Spec.Throughput.GetIngressByteRate(),
		Egress:      quota.Spec.Throughput.GetEgressByteRate(),
		Principals:  principalsToStringSlice(quota.Spec.GetPrincipals()),
		Cluster:     quota.Spec.Cluster.GetId(),
		Environment: quota.Spec.Environment.GetId(),
	})
	return table.Print()
}

func principalsToStringSlice(principals []kafkaquotasv1.GlobalObjectReference) []string {
	ids := make([]string, len(principals))
	for i, principal := range principals {
		ids[i] = principal.GetId()
	}

	return ids
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
