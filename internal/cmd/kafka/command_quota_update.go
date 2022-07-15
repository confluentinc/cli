package kafka

import (
	kafkaquotas "github.com/confluentinc/ccloud-sdk-go-v2-internal/kafka-quotas/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

func (c *quotaCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a previously created cluster link.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
		Example: examples.BuildExampleString(examples.Example{
			Text: `Add "sa-1234" to an existing quota and remove "sa-5678".'`,
			Code: `confluent kafka quota update cq-123ab --add-principals sa-1234 --remove-principals sa-5678`,
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

	getReq := c.V2Client.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.GetKafkaQuotasV1ClientQuota(c.quotaContext(), quotaId)
	quota, _, err := getReq.Execute()
	if err != nil {
		return quotaErr(err)
	}

	updateName, err := getUpdateName(cmd, *quota.DisplayName)
	if err != nil {
		return err
	}
	updateDescription, err := getUpdateDescription(cmd, *quota.Description)
	if err != nil {
		return err
	}
	updateThroughput, err := getUpdateThroughput(cmd, quota.Throughput)
	if err != nil {
		return err
	}
	updatePrincipals, err := c.getUpdatePrincipals(cmd, quota.Principals)
	if err != nil {
		return err
	}

	updatedQuota, err := c.V2Client.CreateKafkaQuota(updateName, updateDescription, updateThroughput,
		&kafkaquotas.ObjectReference{Id: quota.Cluster.Id}, updatePrincipals,
		&kafkaquotas.ObjectReference{Id: quota.Environment.Id},
	)
	if err != nil {
		return quotaErr(err)
	}
	printableQuota := quotaToPrintable(updatedQuota)
	return output.DescribeObject(cmd, printableQuota, quotaListFields, humanRenames, structuredRenames)
}

func (c *quotaCommand) getUpdatePrincipals(cmd *cobra.Command, principals *[]kafkaquotas.ObjectReference) (*[]kafkaquotas.ObjectReference, error) {
	updatePrincipals := *principals
	if cmd.Flags().Changed("add-principals") {
		serviceAccountsToAdd, err := cmd.Flags().GetStringSlice("add-principals")
		if err != nil {
			return &updatePrincipals, err
		}
		principalsToAdd := c.sliceToObjRefArray(serviceAccountsToAdd)
		updatePrincipals = append(updatePrincipals, *principalsToAdd...)
	}
	if cmd.Flags().Changed("remove-principals") {
		principalsToRemove, err := cmd.Flags().GetStringSlice("remove-principals")
		if err != nil {
			return &updatePrincipals, err
		}
		// TODO on upgrage to Go 1.18+ -- instead of using map just do slices.Contains()
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

func getUpdateThroughput(cmd *cobra.Command, throughput *kafkaquotas.KafkaQuotasV1Throughput) (*kafkaquotas.KafkaQuotasV1Throughput, error) {
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

func getUpdateDescription(cmd *cobra.Command, description string) (string, error) {
	if cmd.Flags().Changed("description") {
		updatedDescription, err := cmd.Flags().GetString("description")
		return updatedDescription, err
	}
	return description, nil
}

func getUpdateName(cmd *cobra.Command, name string) (string, error) {
	if cmd.Flags().Changed("name") {
		updatedName, err := cmd.Flags().GetString("name")
		return updatedName, err
	}
	return name, nil
}
