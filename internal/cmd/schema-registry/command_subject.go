package schemaregistry

import (
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

type subjectCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	srClient *srsdk.APIClient
}

func newSubjectCommand(cfg *v1.Config, prerunner pcmd.PreRunner, srClient *srsdk.APIClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "subject",
		Short:       "Manage Schema Registry subjects.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	c := &subjectCommand{srClient: srClient}

	if cfg.IsCloudLogin() {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)
		c.AddCommand(c.newDescribeCommand())
		c.AddCommand(c.newListCommand())
		c.AddCommand(c.newUpdateCommand())
	} else {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)
		c.AddCommand(c.newDescribeCommandOnPrem())
		c.AddCommand(c.newListCommandOnPrem())
		c.AddCommand(c.newUpdateCommandOnPrem())
	}

	return c.Command
}
