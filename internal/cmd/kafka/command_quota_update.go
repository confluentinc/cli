package kafka

import (
	kafkaquotas "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *quotaCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a previously created cluster link.",
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
		return quotaErr(err)
	}

	updateName, err := getUpdatedName(cmd, *quota.DisplayName)
	if err != nil {
		return err
	}
	updateDescription, err := getUpdatedDescription(cmd, *quota.Description)
	if err != nil {
		return err
	}
	updateThroughput, err := getUpdatedThroughput(cmd, quota.Throughput)
	if err != nil {
		return err
	}
	updatePrincipals, err := c.getUpdatedPrincipals(cmd, quota.Principals)
	if err != nil {
		return err
	}

	updatedQuota, err := c.V2Client.UpdateKafkaQuota(kafkaquotas.KafkaQuotasV1ClientQuotaUpdate{
		Id:          &quotaId,
		DisplayName: &updateName,
		Description: &updateDescription,
		Throughput:  updateThroughput,
		Principals:  updatePrincipals,
	})
	if err != nil {
		return quotaErr(err)
	}
	// add back cluster and environment for printing
	updatedQuota.Cluster = quota.Cluster
	updatedQuota.Environment = quota.Environment
	format, _ := cmd.Flags().GetString(output.FlagName)
	printableQuota := quotaToPrintable(updatedQuota, format)
	return output.DescribeObject(cmd, printableQuota, quotaListFields, humanRenames, structuredRenames)
}

func (c *quotaCommand) getUpdatedPrincipals(cmd *cobra.Command, principals *[]kafkaquotas.ObjectReference) (*[]kafkaquotas.ObjectReference, error) {
	updatePrincipals := *principals
	if cmd.Flags().Changed("add-principals") {
		serviceAccountsToAdd, err := cmd.Flags().GetStringSlice("add-principals")
		if err != nil {
			return nil, err
		}
		principalsToAdd := c.sliceToObjRefArray(serviceAccountsToAdd)
		updatePrincipals = append(updatePrincipals, *principalsToAdd...)
	}
	if cmd.Flags().Changed("remove-principals") {
		principalsToRemove, err := cmd.Flags().GetStringSlice("remove-principals")
		if err != nil {
			return nil, err
		}
		// TODO on upgrade to Go 1.18+ -- instead of using map just do slices.Contains()
		removePrincipalMap := make(map[string]struct{})
		for _, p := range principalsToRemove {
			removePrincipalMap[p] = struct{}{}
		}
		i := 0
		for _, principal := range updatePrincipals {
			if _, ok := removePrincipalMap[principal.Id]; !ok {
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
		updatedName, err := cmd.Flags().GetString("name")
		return updatedName, err
	}
	return name, nil
}
