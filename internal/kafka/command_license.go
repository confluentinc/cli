package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type licenseCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newLicenseCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "license",
		Short:       "Manage Confluent Platform licenses.",
		Long:        "Manage Confluent Platform licenses on a Kafka cluster.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	c := &licenseCommand{pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)}
	cmd.PersistentPreRunE = prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand)

	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}
