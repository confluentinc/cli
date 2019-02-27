package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/command"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/mock"
	"github.com/confluentinc/cli/shared"
	cliVersion "github.com/confluentinc/cli/version"
)

func TestAddCommands_MissingPluginsNotShownInHelpUsage(t *testing.T) {
	req := require.New(t)

	logger := log.New()
	cfg := shared.NewConfig(&shared.Config{
		Logger: logger,
	})

	version := cliVersion.NewVersion("1.2.3", "abc1234", "01/23/45", "CI")
	factory := &mock.GRPCPluginFactory{
		CreateFunc: func(name string) common.GRPCPlugin {
			return &mock.GRPCPlugin{
				LookupPathFunc: func() (s string, e error) {
					// return an error to show the plugin wasn't "found" and isn't available
					return "", fmt.Errorf("nada")
				},
			}
		},
	}
	root := BuildCommand(cfg, version, factory, logger)
	prompt := command.NewTerminalPrompt(os.Stdin)
	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		prompt.SetOutput(cmd.OutOrStderr())
	}

	output, err := command.ExecuteCommand(root, "help")
	req.NoError(err)
	req.NotContains(output, "kafka")
	req.NotContains(output, "connect")
	req.NotContains(output, "ksql")
}

func TestAddCommands_AvailablePluginsShownInHelpUsage(t *testing.T) {
	req := require.New(t)

	logger := log.New()
	cfg := shared.NewConfig(&shared.Config{
		Logger: logger,
	})

	version := cliVersion.NewVersion("1.2.3", "abc1234", "01/23/45", "CI")
	factory := &mock.GRPCPluginFactory{
		CreateFunc: func(name string) common.GRPCPlugin {
			return &mock.GRPCPlugin{
				LookupPathFunc: func() (s string, e error) {
					// as long as we don't return an error, the plugin is "found" and available
					return "", nil
				},
			}
		},
	}
	root := BuildCommand(cfg, version, factory, logger)
	prompt := command.NewTerminalPrompt(os.Stdin)
	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		prompt.SetOutput(cmd.OutOrStderr())
	}

	output, err := command.ExecuteCommand(root, "help")
	req.NoError(err)
	req.Contains(output, "kafka")
	req.Contains(output, "connect")
	req.Contains(output, "ksql")
}
