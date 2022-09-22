package kafka

import (
	"fmt"

	kafkaquotasv1 "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	humanRenames      = map[string]string{"Id": "ID", "DisplayName": "Name"}
	structuredRenames = map[string]string{"Id": "id", "DisplayName": "name", "Description": "description", "Ingress": "ingress", "Egress": "egress", "Throughput": "throughput", "Cluster": "cluster", "Principals": "principals", "Environment": "environment"}
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
			}),
	}

	cmd.Flags().String("name", "", "Display name for quota.")
	cmd.Flags().String("description", "", "Description of quota.")
	cmd.Flags().String("ingress", "", "Ingress throughput limit for client (bytes/second).")
	cmd.Flags().String("egress", "", "Egress throughput limit for client (bytes/second).")
	cmd.Flags().StringSlice("principals", []string{}, `A comma delimited list of service accounts to apply quota to. Use "<default>" to apply quota to all service accounts.`)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func (c *quotaCommand) create(cmd *cobra.Command, _ []string) error {
	serviceAccounts, err := cmd.Flags().GetStringSlice("principals")
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
	if err != nil {
		return err
	}

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}
	quotaToCreate := kafkaquotasv1.KafkaQuotasV1ClientQuota{
		DisplayName: &displayName,
		Description: &description,
		Throughput:  throughput,
		Cluster:     &kafkaquotasv1.ObjectReference{Id: cluster.ID},
		Principals:  principals,
		Environment: &kafkaquotasv1.ObjectReference{Id: c.EnvironmentId()},
	}
	quota, httpResp, err := c.V2Client.CreateKafkaQuota(quotaToCreate)
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}
	format, _ := cmd.Flags().GetString(output.FlagName)
	printableQuota := quotaToPrintable(quota, format)
	return output.DescribeObject(cmd, printableQuota, quotaListFields, humanRenames, structuredRenames)
}

func sliceToObjRefArray(accounts []string) *[]kafkaquotasv1.ObjectReference {
	a := make([]kafkaquotasv1.ObjectReference, len(accounts))
	for i := range a {
		a[i] = kafkaquotasv1.ObjectReference{
			Id: accounts[i],
		}
	}
	return &a
}

func getQuotaThroughput(cmd *cobra.Command) (*kafkaquotasv1.KafkaQuotasV1Throughput, error) {
	var throughput kafkaquotasv1.KafkaQuotasV1Throughput

	ingress, err := cmd.Flags().GetString("ingress")
	if err != nil {
		return nil, err
	}
	throughput.IngressByteRate = &ingress

	egress, err := cmd.Flags().GetString("egress")
	if err != nil {
		return nil, err
	}
	throughput.EgressByteRate = &egress

	if ingress == "" || egress == "" {
		return nil, fmt.Errorf(errors.MustSpecifyBothFlagsErrorMsg, "ingress", "egress")
	}
	return &throughput, nil
}
