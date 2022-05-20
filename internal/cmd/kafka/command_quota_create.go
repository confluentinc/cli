package kafka

import (
	"context"
	kafkaquotas "github.com/confluentinc/ccloud-sdk-go-v2-internal/kafka-quotas/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

var (
	humanRenames      = map[string]string{"DisplayName": "Display Name", "Description": "Description", "Throughput": "Throughput", "Cluster": "Cluster", "Principals": "Principals"}
	structuredRenames = map[string]string{"DisplayName": "display_name", "Description": "description", "Throughput": "throughput", "Cluster": "cluster", "Principals": "principals"}
)

func (c *quotaCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Kafka client quota.",
		Args:  cobra.NoArgs,
		RunE:  c.create,
	}

	cmd.Flags().String("ingress", "", "Ingress throughput limit for client.")
	cmd.Flags().String("egress", "", "Egress throughput limit for client.")
	cmd.Flags().String("description", "", "Description of quota.")
	cmd.Flags().String("name", "", "Display name for quota.")
	cmd.Flags().StringSlice("service-accounts", []string{}, "List of service accounts to apply quota for (comma-separated).")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *quotaCommand) create(cmd *cobra.Command, _ []string) error {
	serviceAccounts, err := cmd.Flags().GetStringSlice("service-accounts")
	if err != nil {
		return err
	}
	principals := sliceToObjRefArray(serviceAccounts)

	displayName, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	throughput, err := getQuotaThroughput(cmd)

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	req := c.V2Client.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.CreateKafkaQuotasV1ClientQuota(context.Background())
	envId := c.EnvironmentId()
	req = req.KafkaQuotasV1ClientQuota(kafkaquotas.KafkaQuotasV1ClientQuota{
		DisplayName: &displayName,
		Description: &description,
		Throughput:  throughput,
		Cluster:     &kafkaquotas.ObjectReference{Id: cluster.ID, Environment: &envId},
		Principals:  principals,
	})
	quota, _, err := req.Execute()
	if err != nil {
		return err
	}
	return output.DescribeObject(cmd, quota, quotaListFields, humanRenames, structuredRenames)
}

func sliceToObjRefArray(accounts []string) *[]kafkaquotas.ObjectReference {
	a := make([]kafkaquotas.ObjectReference, len(accounts))
	for i := range a {
		a[i] = kafkaquotas.ObjectReference{
			Id: accounts[i],
		}
	}
	return &a
}

func getQuotaThroughput(cmd *cobra.Command) (*kafkaquotas.KafkaQuotasV1Throughput, error) {
	if cmd.Flags().Changed("ingress") && cmd.Flags().Changed("egress") {
		return nil, errors.New(errors.OnlySpecifyIngressOrEgress)
	}
	var throughput kafkaquotas.KafkaQuotasV1Throughput
	if cmd.Flags().Changed("ingress") {
		ingress, err := cmd.Flags().GetString("ingress")
		if err != nil {
			return nil, err
		}
		throughput.IngressByteRate = &ingress
	} else if cmd.Flags().Changed("egress") {
		egress, err := cmd.Flags().GetString("egress")
		if err != nil {
			return nil, err
		}
		throughput.EgressByteRate = &egress
	}
	return &throughput, nil
}
