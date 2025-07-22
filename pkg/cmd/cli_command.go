package cmd

import (
	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/flink"
	"github.com/confluentinc/cli/v4/pkg/version"
)

type CLICommand struct {
	*cobra.Command
	Config  *config.Config
	Version *version.Version

	cmfClient *flink.CmfRestClient
}

func NewCLICommand(cmd *cobra.Command) *CLICommand {
	return &CLICommand{
		Command: cmd,
		Config:  &config.Config{},
	}
}

func NewAnonymousCLICommand(cmd *cobra.Command, prerunner PreRunner) *CLICommand {
	c := NewCLICommand(cmd)
	cmd.PersistentPreRunE = chain(prerunner.Anonymous(c, false), prerunner.ParseFlagsIntoContext(c))
	return c
}

func (c *CLICommand) GetCmfClient(cmd *cobra.Command) (*flink.CmfRestClient, error) {
	if c.cmfClient != nil {
		return c.cmfClient, nil
	}

	cfg := cmfsdk.NewConfiguration()

	// Set debug flag and user-agent
	unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
	if err != nil {
		return nil, err
	}
	cfg.Debug = unsafeTrace
	cfg.UserAgent = c.Version.UserAgent

	// Resolve flags
	restFlags, err := flink.ResolveOnPremCmfRestFlags(cmd)
	if err != nil {
		return nil, err
	}

	// Construct CMF client (URL logic fully handled inside NewCmfRestClient)
	c.cmfClient, err = flink.NewCmfRestClient(cfg, restFlags, c.Config.IsTest)
	if err != nil {
		return nil, err
	}

	return c.cmfClient, nil
}
