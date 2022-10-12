package kafka

import (
	kafkaquotas "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/set"
)

func (c *quotaCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a Kafka client quota.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
		Example: examples.BuildExampleString(examples.Example{
			Text: `Add "sa-12345" to an existing quota and remove "sa-67890".`,
			Code: `confluent kafka quota update cq-123ab --add-principals sa-12345 --remove-principals sa-67890`,
		}),
	}

	cmd.Flags().String("ingress", "", "Update ingress limit for quota.")
	cmd.Flags().String("egress", "", "Update egress limit for quota.")
	cmd.Flags().StringSlice("add-principals", []string{}, "List of service accounts to add to quota (comma-separated).")
	cmd.Flags().StringSlice("remove-principals", []string{}, "List of service accounts to remove from quota (comma-separated).")
	cmd.Flags().String("description", "", "Update description.")
	cmd.Flags().String("name", "", "Update display name.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *quotaCommand) update(cmd *cobra.Command, args []string) error {
	quotaId := args[0]

	quota, err := c.V2Client.DescribeKafkaQuota(quotaId)
	if err != nil {
		return err
	}

	updateName, err := getUpdatedName(cmd, *quota.Spec.DisplayName)
	if err != nil {
		return err
	}
	updateDescription, err := getUpdatedDescription(cmd, *quota.Spec.Description)
	if err != nil {
		return err
	}
	updateThroughput, err := getUpdatedThroughput(cmd, quota.Spec.Throughput)
	if err != nil {
		return err
	}
	updatePrincipals, err := c.getUpdatedPrincipals(cmd, *quota.Spec.Principals)
	if err != nil {
		return err
	}
	quotaUpdate := kafkaquotas.KafkaQuotasV1ClientQuotaUpdate{
		Id: &quotaId,
		Spec: &kafkaquotas.KafkaQuotasV1ClientQuotaSpecUpdate{
			DisplayName: &updateName,
			Description: &updateDescription,
			Throughput:  updateThroughput,
			Principals:  updatePrincipals,
		},
	}
	updatedQuota, err := c.V2Client.UpdateKafkaQuota(quotaUpdate)
	if err != nil {
		return err
	}
	format, _ := cmd.Flags().GetString(output.FlagName)
	printQuota := quotaToPrintable(updatedQuota, format)
	return output.DescribeObject(cmd, printQuota, quotaListFields, humanRenames, structuredRenames)
}

func (c *quotaCommand) getUpdatedPrincipals(cmd *cobra.Command, updatePrincipals []kafkaquotas.GlobalObjectReference) (*[]kafkaquotas.GlobalObjectReference, error) {
	if cmd.Flags().Changed("add-principals") {
		serviceAccountsToAdd, err := cmd.Flags().GetStringSlice("add-principals")
		if err != nil {
			return nil, err
		}
		principalsToAdd := sliceToObjRefArray(serviceAccountsToAdd)
		updatePrincipals = append(updatePrincipals, *principalsToAdd...)
	}
	if cmd.Flags().Changed("remove-principals") {
		principalsToRemove, err := cmd.Flags().GetStringSlice("remove-principals")
		if err != nil {
			return nil, err
		}
		removePrincipals := set.New()
		for _, p := range principalsToRemove {
			removePrincipals.Add(p)
		}
		i := 0
		for _, principal := range updatePrincipals {
			if contains := removePrincipals.Contains(principal.Id); !contains {
				updatePrincipals[i] = principal
				i++
			}
		}
		updatePrincipals = updatePrincipals[:i]
	}
	return &updatePrincipals, nil
}

func getUpdatedThroughput(cmd *cobra.Command, throughput *kafkaquotas.KafkaQuotasV1Throughput) (*kafkaquotas.KafkaQuotasV1Throughput, error) {
	if cmd.Flags().Changed("ingress") {
		updatedIngress, err := cmd.Flags().GetString("ingress")
		if err != nil {
			return throughput, err
		}
		throughput.SetIngressByteRate(updatedIngress)
	}
	if cmd.Flags().Changed("egress") {
		updatedEgress, err := cmd.Flags().GetString("egress")
		if err != nil {
			return throughput, err
		}
		throughput.SetEgressByteRate(updatedEgress)
	}
	return throughput, nil
}

func getUpdatedDescription(cmd *cobra.Command, description string) (string, error) {
	if cmd.Flags().Changed("description") {
		return cmd.Flags().GetString("description")
	}
	return description, nil
}

func getUpdatedName(cmd *cobra.Command, name string) (string, error) {
	if cmd.Flags().Changed("name") {
		return cmd.Flags().GetString("name")
	}
	return name, nil
}
