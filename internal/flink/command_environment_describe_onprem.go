package flink

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newEnvironmentDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.environmentDescribe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) environmentDescribe(cmd *cobra.Command, args []string) error {
	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	// Get the name of the environment to be retrieved
	environmentName := args[0]
	cmfEnvironment, httpResponse, err := cmfClient.DefaultApi.GetEnvironment(cmd.Context(), environmentName)

	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return fmt.Errorf(`failed to describe environment "%s": %s`, environmentName, parsedErr)
	}

	table := output.NewTable(cmd)
	var defaultsBytes []byte
	defaultsBytes, err = json.Marshal(cmfEnvironment.Defaults)
	if err != nil {
		return fmt.Errorf(`failed to marshal defaults for environment "%s": %s`, environmentName, err)
	}

	table.Add(&flinkEnvironmentOutput{
		Name:        cmfEnvironment.Name,
		Defaults:    string(defaultsBytes),
		CreatedTime: cmfEnvironment.CreatedTime.String(),
		UpdatedTime: cmfEnvironment.UpdatedTime.String(),
	})
	return table.Print()
}
