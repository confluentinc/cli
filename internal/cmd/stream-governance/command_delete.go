package streamgovernance

import (
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/cli/internal/pkg/version"
	"github.com/spf13/cobra"
	"os"
)

func (c *streamGovernanceCommand) newDeleteCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete the Stream Governance cluster for an environment.",
		Args:  cobra.NoArgs,
		RunE: pcmd.NewCLIRunE(func(cmd *cobra.Command, args []string) error {
			return c.delete(cmd, args, form.NewPrompt(os.Stdin))
		}),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete Stream Governance cluster for environment 'env-00000'",
				Code: fmt.Sprintf("%s stream-governance delete --environment env-00000", version.CLIName),
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *streamGovernanceCommand) delete(cmd *cobra.Command, _ []string, prompt form.Prompt) error {
	ctx := c.V2Client.StreamGovernanceApiContext()

	clusterId, err := c.getStreamGovernanceV2ClusterIdForEnvironment(ctx)
	if err != nil {
		return err
	}

	isDeleteConfirmed, err := deleteConfirmation(cmd, c.EnvironmentId(), prompt)
	if err != nil {
		return err
	}

	if !isDeleteConfirmed {
		utils.Println(cmd, "Terminating operation ...")
		return nil
	}

	_, err = c.V2Client.StreamGovernanceClient.ClustersStreamGovernanceV2Api.DeleteStreamGovernanceV2Cluster(ctx, clusterId).Execute()
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.StreamGovernanceClusterDeletedMsg, c.EnvironmentId())
	return nil
}

func deleteConfirmation(cmd *cobra.Command, environmentId string, prompt form.Prompt) (bool, error) {
	f := form.New(
		form.Field{ID: "confirmation", Prompt: fmt.Sprintf("Are you sure you want to delete the Stream Governance "+
			"cluster for environment %s?", environmentId), IsYesOrNo: true},
	)
	if err := f.Prompt(cmd, prompt); err != nil {
		return false, errors.New(errors.SGFailedToReadDeletionConfirmationErrorMsg)
	}
	return f.Responses["confirmation"].(bool), nil
}
