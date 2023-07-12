package kafka

import (
	"github.com/spf13/cobra"

	kafkaquotasv1 "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *quotaCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Kafka client quota.",
		Args:  cobra.NoArgs,
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create client quotas for service accounts "sa-1234" and "sa-5678" on cluster "lkc-1234".`,
				Code: `confluent kafka quota create --name clientQuota --ingress 500 --egress 100 --principals sa-1234,sa-5678 --cluster lkc-1234`,
			},
			examples.Example{
				Text: `Create a default client quota for all principals without an explicit quota assignment.`,
				Code: `confluent kafka quota create --name defaultQuota --ingress 500 --egress 500 --principals "<default>" --cluster lkc-1234`,
			},
		),
	}

	cmd.Flags().String("name", "", "Display name for quota.")
	cmd.Flags().String("description", "", "Description of quota.")
	cmd.Flags().String("ingress", "", "Ingress throughput limit for client (bytes/second).")
	cmd.Flags().String("egress", "", "Egress throughput limit for client (bytes/second).")
	cmd.Flags().StringSlice("principals", []string{}, `A comma-separated list of service accounts to apply the quota to. Use "<default>" to apply the quota to all service accounts.`)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("name"))
	cmd.MarkFlagsRequiredTogether("ingress", "egress")

	return cmd
}

func (c *quotaCommand) create(cmd *cobra.Command, _ []string) error {
	serviceAccounts, err := cmd.Flags().GetStringSlice("principals")
	if err != nil {
		return err
	}
	principals := sliceToObjRefArray(serviceAccounts)

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	throughput, err := getQuotaThroughput(cmd)
	if err != nil {
		return err
	}

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	quotaToCreate := kafkaquotasv1.KafkaQuotasV1ClientQuota{
		Spec: &kafkaquotasv1.KafkaQuotasV1ClientQuotaSpec{
			DisplayName: kafkaquotasv1.PtrString(name),
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

	table := output.NewTable(cmd)
	format := output.GetFormat(cmd)
	table.Add(quotaToPrintable(quota, format))
	return table.Print()
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
