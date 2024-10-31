package customcodelogging

import (
	"github.com/spf13/cobra"

	cclv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccl/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
)

type customCodeLoggingCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type customCodeLoggingOut struct {
	Id          string `human:"Id" serialized:"id"`
	Cloud       string `human:"Cloud" serialized:"cloud"`
	Region      string `human:"Region" serialized:"region"`
	Environment string `human:"Environment" serialized:"environment"`
	Destination string `human:"Destination" serialized:"destination"`
	Topic       string `human:"Topic" serialized:"topic"`
	Cluster     string `human:"Cluster" serialized:"cluster"`
	LogLevel    string `human:"Log Level" serialized:"log_level"`
}

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "custom-code-logging",
		Short:       "Manage custom code logging.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &customCodeLoggingCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	_ = cfg.ParseFlagsIntoConfig(cmd)
	if cfg.IsTest || featureflags.Manager.BoolVariation("cli.custom_code_logging.early_access", cfg.Context(), config.CliLaunchDarklyClient, true, false) {
		cmd.AddCommand(c.newCreateCommand())
		cmd.AddCommand(c.newDeleteCommand())
		cmd.AddCommand(c.newDescribeCommand())
		cmd.AddCommand(c.newListCommand())
		cmd.AddCommand(c.newUpdateCommand())
	}

	return cmd
}

func getCustomCodeLogging(customCodeLogging cclv1.CclV1CustomCodeLogging) *customCodeLoggingOut {
	customCodeLoggingOut := &customCodeLoggingOut{
		Id:          customCodeLogging.GetId(),
		Cloud:       customCodeLogging.GetCloud(),
		Region:      customCodeLogging.GetRegion(),
		Environment: customCodeLogging.GetEnvironment().Id,
	}
	if customCodeLogging.GetDestinationSettings().CclV1KafkaDestinationSettings != nil {
		customCodeLoggingOut.Destination = customCodeLogging.GetDestinationSettings().CclV1KafkaDestinationSettings.GetKind()
		customCodeLoggingOut.Topic = customCodeLogging.GetDestinationSettings().CclV1KafkaDestinationSettings.GetTopic()
		customCodeLoggingOut.Cluster = customCodeLogging.GetDestinationSettings().CclV1KafkaDestinationSettings.GetClusterId()
		customCodeLoggingOut.LogLevel = customCodeLogging.GetDestinationSettings().CclV1KafkaDestinationSettings.GetLogLevel()
	}
	return customCodeLoggingOut
}
