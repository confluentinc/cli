package kafka

import (
	"fmt"
	"os"
	"strings"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *clusterCommand) newUpdateCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update a Kafka cluster.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.update(cmd, args, form.NewPrompt(os.Stdin))
		},
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Change a cluster's name and expand its CKU count:",
				Code: `confluent kafka cluster update lkc-abc123 --name "Cool Cluster" --cku 3`,
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the Kafka cluster.")
	pcmd.AddAvailabilityFlag(cmd)
	pcmd.AddTypeFlag(cmd)
	cmd.Flags().Uint32("cku", 0, `Number of Confluent Kafka Units. For Kafka clusters of type "dedicated" only. When shrinking a cluster, you must reduce capacity one CKU at a time.`)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) update(cmd *cobra.Command, args []string, prompt form.Prompt) error {
	if !(cmd.Flags().Changed("name") || cmd.Flags().Changed("availability") || cmd.Flags().Changed("type") || cmd.Flags().Changed("cku")) {
		return errors.New("must specify one of `--name`, `--availability`, `--type`, or `--cku`")
	}

	clusterId := args[0]
	currentCluster, _, err := c.V2Client.DescribeKafkaCluster(clusterId, c.EnvironmentId())
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.KafkaClusterNotFoundErrorMsg, clusterId), errors.ChooseRightEnvironmentSuggestions)
	}

	if currentCluster.Status.GetPhase() == ccloudv2.StatusProvisioning {
		return errors.New(errors.KafkaClusterProvisioningErrorMsg)
	}

	update := cmkv2.CmkV2ClusterUpdate{
		Id:   cmkv2.PtrString(clusterId),
		Spec: &cmkv2.CmkV2ClusterSpecUpdate{Environment: &cmkv2.EnvScopedObjectReference{Id: c.EnvironmentId()}},
	}

	if cmd.Flags().Changed("name") {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		if name == "" {
			return errors.New("`--name` must not be empty")
		}

		update.Spec.SetDisplayName(name)

		// The backend blocks simultaneous modification of `--name` and `--type`. If both are passed, update the name separately.
		if cmd.Flags().Changed("type") {
			if _, err := c.V2Client.UpdateKafkaCluster(clusterId, update); err != nil {
				return err
			}
			update.Spec.DisplayName = nil
		}
	}

	if cmd.Flags().Changed("availability") {
		availability, err := cmd.Flags().GetString("availability")
		if err != nil {
			return err
		}
		if availability == "" {
			return errors.New("`--availability` must not be empty")
		}
		update.Spec.SetAvailability(strings.ToUpper(strings.ReplaceAll(availability, "-", "_")))
	}

	// Both `--type` and `--cku` require changing the config
	if cmd.Flags().Changed("type") || cmd.Flags().Changed("cku") {
		// Set base config before setting type or CKU
		updatedClusterType := getClusterType(currentCluster.Spec.GetConfig())
		if cmd.Flags().Changed("type") {
			clusterType, err := cmd.Flags().GetString("type")
			if err != nil {
				return err
			}
			updatedClusterType = clusterType
		}
		switch updatedClusterType {
		case skuBasic:
			update.Spec.SetConfig(cmkv2.CmkV2ClusterSpecUpdateConfigOneOf{CmkV2Basic: new(cmkv2.CmkV2Basic)})
		case skuStandard:
			update.Spec.SetConfig(cmkv2.CmkV2ClusterSpecUpdateConfigOneOf{CmkV2Standard: new(cmkv2.CmkV2Standard)})
		case skuDedicated:
			update.Spec.SetConfig(cmkv2.CmkV2ClusterSpecUpdateConfigOneOf{CmkV2Dedicated: new(cmkv2.CmkV2Dedicated)})
		default:
			return fmt.Errorf(`unsupported cluster type "%s"`, updatedClusterType)
		}

		if cmd.Flags().Changed("type") {
			clusterType, err := cmd.Flags().GetString("type")
			if err != nil {
				return err
			}
			if clusterType == "" {
				return errors.New("`--type` must not be empty")
			}
			switch clusterType {
			case skuBasic:
				update.Spec.Config.CmkV2Basic.SetKind("Basic")
			case skuStandard:
				update.Spec.Config.CmkV2Standard.SetKind("Standard")
			case skuDedicated:
				update.Spec.Config.CmkV2Dedicated.SetKind("Dedicated")
			}
		}

		if cmd.Flags().Changed("cku") {
			cku, err := cmd.Flags().GetUint32("cku")
			if err != nil {
				return err
			}
			switch updatedClusterType {
			case skuBasic, skuStandard:
				return fmt.Errorf(errors.ClusterResizeNotSupportedErrorMsg)
			case skuDedicated:
				updatedCku, err := c.validateResize(int32(cku), &currentCluster, prompt)
				if err != nil {
					return err
				}
				update.Spec.Config.CmkV2Dedicated.SetCku(updatedCku)
			}
		}
	}

	updatedCluster, err := c.V2Client.UpdateKafkaCluster(clusterId, update)
	if err != nil {
		return err
	}

	ctx := c.Context.Config.Context()
	c.Context.Config.SetOverwrittenActiveKafka(ctx.KafkaClusterContext.GetActiveKafkaClusterId())
	ctx.KafkaClusterContext.SetActiveKafkaCluster(clusterId)

	return c.outputKafkaClusterDescription(cmd, &updatedCluster, true)
}

func (c *clusterCommand) validateResize(cku int32, currentCluster *cmkv2.CmkV2Cluster, prompt form.Prompt) (int32, error) {
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
	if err := isClusterResizeInProgress(currentCluster); err != nil {
		return 0, err
	}

	// If shrink
	if cku < currentCluster.Spec.GetConfig().CmkV2Dedicated.GetCku() {
		promptMessage := ""
		if err := c.validateKafkaClusterMetrics(currentCluster, true); err != nil {
			promptMessage += fmt.Sprintf("\n%v\n", err)
		}
		if err := c.validateKafkaClusterMetrics(currentCluster, false); err != nil {
			promptMessage += fmt.Sprintf("\n%v\n", err)
		}
		if promptMessage != "" {
			if ok, err := confirmShrink(prompt, promptMessage); !ok || err != nil {
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

func confirmShrink(prompt form.Prompt, promptMessage string) (bool, error) {
	f := form.New(form.Field{ID: "proceed", Prompt: fmt.Sprintf("Validated cluster metrics and found that: %s\nDo you want to proceed with shrinking your kafka cluster?", promptMessage), IsYesOrNo: true})
	if err := f.Prompt(prompt); err != nil {
		return false, errors.New(errors.FailedToReadClusterResizeConfirmationErrorMsg)
	}
	if !f.Responses["proceed"].(bool) {
		output.Println("Not proceeding with kafka cluster shrink")
		return false, nil
	}
	return true, nil
}

func getClusterType(config cmkv2.CmkV2ClusterSpecConfigOneOf) string {
	if config.CmkV2Dedicated != nil {
		return skuDedicated
	}

	if config.CmkV2Standard != nil {
		return skuStandard
	}

	return skuBasic
}
