package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newEnvironmentUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a Flink environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.environmentUpdate,
	}

	addCmfFlagSet(cmd)
	cmd.Flags().String("defaults", "", "JSON string defining the environment's Flink application defaults, or path to a file to read defaults from (with .yml, .yaml or .json extension).")
	cmd.Flags().String("statement-defaults", "", "JSON string defining the environment's Flink statement defaults, or path to a file to read defaults from (with .yml, .yaml or .json extension).")
	cmd.Flags().String("compute-pool-defaults", "", "JSON string defining the environment's Flink compute pool defaults, or path to a file to read defaults from (with .yml, .yaml or .json extension).")
	pcmd.AddOutputFlag(cmd)

	// At least one of the defaults flags must be provided in order to update the environment
	cmd.MarkFlagsOneRequired("defaults", "statement-defaults", "compute-pool-defaults")

	return cmd
}

func (c *command) environmentUpdate(cmd *cobra.Command, args []string) error {
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environmentName := args[0]

	// Read file contents or parse defaults if applicable
	var defaultsApplicationParsed, defaultsComputePoolParsed map[string]interface{}
	var defaultsStatementParsed cmfsdk.AllStatementDefaults1

	defaultsApplication, err := cmd.Flags().GetString("defaults")
	if err != nil {
		return fmt.Errorf("failed to read defaults application: %v", err)
	}

	defaultsComputePool, err := cmd.Flags().GetString("compute-pool-defaults")
	if err != nil {
		return fmt.Errorf("failed to read defaults compute pool: %v", err)
	}

	defaultsStatement, err := cmd.Flags().GetString("statement-defaults")
	if err != nil {
		return fmt.Errorf("failed to read defaults statement: %v", err)
	}

	if defaultsApplication != "" {
		if defaultsApplicationParsed, err = parseDefaultsAsGenericType[map[string]interface{}](defaultsApplication, "application"); err != nil {
			return err
		}
	}
	if defaultsComputePool != "" {
		if defaultsComputePoolParsed, err = parseDefaultsAsGenericType[map[string]interface{}](defaultsComputePool, "compute-pool"); err != nil {
			return err
		}
	}
	if defaultsStatement != "" {
		defaultsStatementParsedLocal, err := parseDefaultsAsGenericType[LocalAllStatementDefaults1](defaultsStatement, "statement")
		if err != nil {
			return err
		}
		if defaultsStatementParsedLocal.Detached != nil {
			defaultsStatementParsed.Detached = &cmfsdk.StatementDefaults{FlinkConfiguration: defaultsStatementParsedLocal.Detached.FlinkConfiguration}
		}
		if defaultsStatementParsedLocal.Interactive != nil {
			defaultsStatementParsed.Interactive = &cmfsdk.StatementDefaults{FlinkConfiguration: defaultsStatementParsedLocal.Interactive.FlinkConfiguration}
		}
	}

	var postEnvironment cmfsdk.PostEnvironment
	postEnvironment.Name = environmentName
	postEnvironment.FlinkApplicationDefaults = &defaultsApplicationParsed
	postEnvironment.StatementDefaults = &defaultsStatementParsed
	postEnvironment.ComputePoolDefaults = &defaultsComputePoolParsed

	sdkOutputEnvironment, err := client.UpdateEnvironment(c.createContext(), postEnvironment)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		return printEnvironmentOutTable(cmd, sdkOutputEnvironment)
	}

	localEnv := convertSdkEnvironmentToLocalEnvironment(sdkOutputEnvironment)
	return output.SerializedOutput(cmd, localEnv)
}
