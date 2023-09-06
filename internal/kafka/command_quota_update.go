package kafka

import (
	"github.com/spf13/cobra"

	kafkaquotasv1 "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/types"
)

func (c *quotaCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update a Kafka client quota.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
		Example: examples.BuildExampleString(examples.Example{
			Text: `Add "sa-12345" to an existing quota and remove "sa-67890".`,
			Code: `confluent kafka quota update cq-123ab --add-principals sa-12345 --remove-principals sa-67890`,
		}),
	}

	cmd.Flags().String("ingress", "", "Update ingress limit for quota.")
	cmd.Flags().String("egress", "", "Update egress limit for quota.")
	cmd.Flags().StringSlice("add-principals", []string{}, "A comma-separated list of service accounts to add to the quota.")
	cmd.Flags().StringSlice("remove-principals", []string{}, "A comma-separated list of service accounts to remove from the quota.")
	cmd.Flags().String("description", "", "Update description.")
	cmd.Flags().String("name", "", "Update display name.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *quotaCommand) update(cmd *cobra.Command, args []string) error {
	flags := []string{
		"add-principals",
		"description",
		"egress",
		"ingress",
		"name",
		"remove-principals",
	}
	if err := errors.CheckNoUpdate(cmd.Flags(), flags...); err != nil {
		return err
	}
	quotaId := args[0]

	quota, err := c.V2Client.DescribeKafkaQuota(quotaId)
	if err != nil {
		return err
	}

	updateName, err := getUpdatedName(cmd, quota.Spec.GetDisplayName())
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

	quotaUpdate := kafkaquotasv1.KafkaQuotasV1ClientQuotaUpdate{
		Id: &quotaId,
		Spec: &kafkaquotasv1.KafkaQuotasV1ClientQuotaSpecUpdate{
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

	table := output.NewTable(cmd)
	format := output.GetFormat(cmd)
	table.Add(quotaToPrintable(updatedQuota, format))
	return table.Print()
}

func (c *quotaCommand) getUpdatedPrincipals(cmd *cobra.Command, updatePrincipals []kafkaquotasv1.GlobalObjectReference) (*[]kafkaquotasv1.GlobalObjectReference, error) {
	if cmd.Flags().Changed("add-principals") {
		addPrincipals, err := cmd.Flags().GetStringSlice("add-principals")
		if err != nil {
			return nil, err
		}
		principalsToAdd := sliceToObjRefArray(addPrincipals)
		updatePrincipals = append(updatePrincipals, *principalsToAdd...)
	}
	if cmd.Flags().Changed("remove-principals") {
		removePrincipals, err := cmd.Flags().GetStringSlice("remove-principals")
		if err != nil {
			return nil, err
		}
		remove := types.NewSet[string]()
		for _, p := range removePrincipals {
			remove.Add(p)
		}
		i := 0
		for _, principal := range updatePrincipals {
			if contains := remove.Contains(principal.Id); !contains {
				updatePrincipals[i] = principal
				i++
			}
		}
		updatePrincipals = updatePrincipals[:i]
	}
	return &updatePrincipals, nil
}

func getUpdatedThroughput(cmd *cobra.Command, throughput *kafkaquotasv1.KafkaQuotasV1Throughput) (*kafkaquotasv1.KafkaQuotasV1Throughput, error) {
	if cmd.Flags().Changed("ingress") {
		ingress, err := cmd.Flags().GetString("ingress")
		if err != nil {
			return throughput, err
		}
		throughput.SetIngressByteRate(ingress)
	}
	if cmd.Flags().Changed("egress") {
		egress, err := cmd.Flags().GetString("egress")
		if err != nil {
			return throughput, err
		}
		throughput.SetEgressByteRate(egress)
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
