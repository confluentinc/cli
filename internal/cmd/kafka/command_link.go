package kafka

// TODO: wrap all link / mirror commands with kafka rest error
import (
	"context"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	configFileFlagName = "config-file"
	dryrunFlagName     = "dry-run"
)

type linkCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	cfg *v1.Config
}

func newLinkCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
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

func (c *linkCommand) getKafkaRestComponents(cmd *cobra.Command) (*kafkarestv3.APIClient, context.Context, string, error) {
	if c.cfg.IsCloudLogin() {
		kafkaRest, err := c.GetKafkaREST()
		if kafkaRest == nil {
			if err != nil {
				return nil, nil, "", err
			}
			return nil, nil, "", errors.New(errors.RestProxyNotAvailableMsg)
		}

		clusterId, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand)
		if err != nil {
			return nil, nil, "", err
		}

		return kafkaRest.Client, kafkaRest.Context, clusterId, nil
	} else {
		client, ctx, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
		if err != nil {
			return nil, nil, "", err
		}

		clusterId, err := getClusterIdForRestRequests(client, ctx)
		if err != nil {
			return nil, nil, "", err
		}

		return client, ctx, clusterId, nil
	}
}
