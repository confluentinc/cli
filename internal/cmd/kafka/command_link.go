package kafka

// TODO: wrap all link / mirror commands with kafka rest error
import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

const (
	configFileFlagName = "config-file"
	dryrunFlagName     = "dry-run"
)

type linkCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	cfg *v1.Config
}

func newLinkCommand(cfg *v1. Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "link",
		Short:       "Manages inter-cluster links.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &linkCommand{cfg: cfg}

	c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)
	if cfg.IsOnPremLogin() {
		c.SetPersistentPreRunE(prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand))
	}

	c.AddCommand(c.newCreateCommand())
	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newUpdateCommand())

	return c.Command
}
