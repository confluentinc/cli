package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	pconv "github.com/confluentinc/cli/internal/pkg/name-conversions"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *clusterCommand) newUpdateCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id|name>",
		Short:             "Update a Kafka cluster.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update the name and CKU count of a Kafka cluster:",
				Code: `confluent kafka cluster update lkc-123456 --name "New Cluster Name" --cku 3`,
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the Kafka cluster.")
	cmd.Flags().Uint32("cku", 0, `Number of Confluent Kafka Units. For Kafka clusters of type "dedicated" only. When shrinking a cluster, you must reduce capacity one CKU at a time.`)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) update(cmd *cobra.Command, args []string) error {
	if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("cku") {
		return errors.New(errors.NameOrCKUFlagErrorMsg)
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	clusterID := args[0]
	currentCluster, _, err := c.V2Client.DescribeKafkaCluster(clusterID, environmentId)
	if err != nil {
		environmentId, err = pconv.ConvertEnvironmentNameToId(environmentId, c.V2Client)
		if err != nil {
			return err
		}
		clusterID, err = pconv.ConvertClusterNameToId(clusterID, environmentId, c.V2Client)
		if err != nil {
			return err
		}
		currentCluster, _, err = c.V2Client.DescribeKafkaCluster(clusterID, environmentId)
		if err != nil {
			return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.KafkaClusterNotFoundErrorMsg, clusterID), errors.ChooseRightEnvironmentSuggestions)
		}
	}

	update := cmkv2.CmkV2ClusterUpdate{
		Id:   cmkv2.PtrString(clusterID),
		Spec: &cmkv2.CmkV2ClusterSpecUpdate{Environment: &cmkv2.EnvScopedObjectReference{Id: environmentId}},
	}

	if cmd.Flags().Changed("name") {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		if name == "" {
			return errors.New(errors.NonEmptyNameErrorMsg)
		}
		update.Spec.SetDisplayName(name)
	}

	if cmd.Flags().Changed("cku") {
		cku, err := cmd.Flags().GetUint32("cku")
		if err != nil {
			return err
		}
		updatedCku, err := c.validateResize(int32(cku), &currentCluster)
		if err != nil {
			return err
		}
		update.Spec.Config = &cmkv2.CmkV2ClusterSpecUpdateConfigOneOf{CmkV2Dedicated: &cmkv2.CmkV2Dedicated{Kind: "Dedicated", Cku: updatedCku}}
	}

	updatedCluster, err := c.V2Client.UpdateKafkaCluster(clusterID, update)
	if err != nil {
		return errors.NewWrapErrorWithSuggestions(err, "failed to update Kafka cluster", errors.KafkaClusterUpdateFailedSuggestions)
	}

	ctx := c.Context.Config.Context()
	c.Context.Config.SetOverwrittenCurrentKafkaCluster(ctx.KafkaClusterContext.GetActiveKafkaClusterId())
	ctx.KafkaClusterContext.SetActiveKafkaCluster(clusterID)

	return c.outputKafkaClusterDescription(cmd, &updatedCluster, true)
}

func (c *clusterCommand) validateResize(cku int32, currentCluster *cmkv2.CmkV2Cluster) (int32, error) {
	// Ensure the cluster is a Dedicated Cluster
	if currentCluster.GetSpec().Config.CmkV2Dedicated == nil {
		return 0, errors.New(errors.ClusterResizeNotSupportedErrorMsg)
	}
	// Durability Checks
	if currentCluster.Spec.GetAvailability() == highAvailability && cku <= 1 {
		return 0, errors.New(errors.CKUMoreThanOneErrorMsg)
	}
	if cku == 0 {
		return 0, errors.New(errors.CKUMoreThanZeroErrorMsg)
	}
	// Cluster can't be resized while it's provisioning or being expanded already.
	// Name _can_ be changed during these times, though.
	if err := isClusterResizeInProgress(currentCluster); err != nil {
		return 0, err
	}
	// If shrink
	if cku < currentCluster.GetSpec().Config.CmkV2Dedicated.Cku {
		promptMessage := ""
		// metrics api auth via jwt
		if err := c.validateKafkaClusterMetrics(currentCluster, true); err != nil {
			promptMessage += fmt.Sprintf("\n%v\n", err)
		}
		if err := c.validateKafkaClusterMetrics(currentCluster, false); err != nil {
			promptMessage += fmt.Sprintf("\n%v\n", err)
		}
		if promptMessage != "" {
			if ok, err := confirmShrink(promptMessage); !ok || err != nil {
				return 0, err
			}
		}
	}
	return cku, nil
}

func (c *clusterCommand) validateKafkaClusterMetrics(currentCluster *cmkv2.CmkV2Cluster, isLatestMetric bool) error {
	window := "3 day"
	if isLatestMetric {
		window = "15 min"
	}

	if err := c.validateClusterLoad(*currentCluster.Id, isLatestMetric); err != nil {
		return errors.Errorf("Looking at metrics in the last %s window:\n%v", window, err)
	}

	return nil
}

func confirmShrink(promptMessage string) (bool, error) {
	f := form.New(form.Field{ID: "proceed", Prompt: fmt.Sprintf("Validated cluster metrics and found that: %s\nDo you want to proceed with shrinking your kafka cluster?", promptMessage), IsYesOrNo: true})
	if err := f.Prompt(form.NewPrompt()); err != nil {
		return false, errors.New(errors.FailedToReadClusterResizeConfirmationErrorMsg)
	}
	if !f.Responses["proceed"].(bool) {
		output.Println("Not proceeding with kafka cluster shrink")
		return false, nil
	}
	return true, nil
}
