package kafka

import (
	"github.com/spf13/cobra"

	kafkaquotasv1 "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafka"
)

func (c *quotaCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Kafka client quota.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create client quotas for service accounts "sa-1234" and "sa-5678" on cluster "lkc-1234".`,
				Code: `confluent kafka quota create my-client-quota --ingress 500 --egress 100 --principals sa-1234,sa-5678 --cluster lkc-1234`,
			},
			examples.Example{
				Text: `Create a default client quota for all principals without an explicit quota assignment.`,
				Code: `confluent kafka quota create my-default-quota --ingress 500 --egress 500 --principals "<default>" --cluster lkc-1234`,
			},
		),
	}

	cmd.Flags().String("description", "", "Description of quota.")
	cmd.Flags().String("ingress", "", "Ingress throughput limit for client (bytes/second).")
	cmd.Flags().String("egress", "", "Egress throughput limit for client (bytes/second).")
	cmd.Flags().StringSlice("principals", []string{}, `A comma-separated list of service accounts to apply the quota to. Use "<default>" to apply the quota to all service accounts.`)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsRequiredTogether("ingress", "egress")

	return cmd
}

func (c *quotaCommand) create(cmd *cobra.Command, args []string) error {
	serviceAccounts, err := cmd.Flags().GetStringSlice("principals")
	if err != nil {
		return err
	}
	principals := sliceToObjRefArray(serviceAccounts)

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	throughput, err := getQuotaThroughput(cmd)
	if err != nil {
		return err
	}

	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	quotaToCreate := kafkaquotasv1.KafkaQuotasV1ClientQuota{
		Spec: &kafkaquotasv1.KafkaQuotasV1ClientQuotaSpec{
			DisplayName: kafkaquotasv1.PtrString(args[0]),
			Description: kafkaquotasv1.PtrString(description),
			Throughput:  throughput,
			Cluster:     &kafkaquotasv1.EnvScopedObjectReference{Id: cluster.ID},
			Principals:  principals,
			Environment: &kafkaquotasv1.GlobalObjectReference{Id: environmentId},
		},
	}

	quota, err := c.V2Client.CreateKafkaQuota(quotaToCreate)
	if err != nil {
		return err
	}

	return printTable(cmd, quota)
}

func sliceToObjRefArray(accounts []string) *[]kafkaquotasv1.GlobalObjectReference {
	a := make([]kafkaquotasv1.GlobalObjectReference, len(accounts))
	for i := range a {
		a[i] = kafkaquotasv1.GlobalObjectReference{
			Id: accounts[i],
		}
	}
	return &a
}

func getQuotaThroughput(cmd *cobra.Command) (*kafkaquotasv1.KafkaQuotasV1Throughput, error) {
	ingress, err := cmd.Flags().GetString("ingress")
	if err != nil {
		return nil, err
	}

	egress, err := cmd.Flags().GetString("egress")
	if err != nil {
		return nil, err
	}

	throughput := &kafkaquotasv1.KafkaQuotasV1Throughput{
		IngressByteRate: ingress,
		EgressByteRate:  egress,
	}

	return throughput, nil
}
