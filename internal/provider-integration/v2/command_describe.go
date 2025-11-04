// Copyright 2024 Confluent Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v2

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a provider integration.",
		Long:  "Describe a provider integration including its configuration and status.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe provider integration "pi-123456" in the current environment.`,
				Code: "confluent provider-integration v2 describe pi-123456",
			},
			examples.Example{
				Text: `Describe provider integration "pi-123456" in environment "env-789012".`,
				Code: "confluent provider-integration v2 describe pi-123456 --environment env-789012",
			},
		),
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	integrationId := args[0]

	environmentId, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	if environmentId == "" {
		environmentId, err = c.Context.EnvironmentId()
		if err != nil {
			return err
		}
	}

	integration, err := c.V2Client.GetPimV2Integration(cmd.Context(), integrationId, environmentId)
	if err != nil {
		return err
	}

	return printProviderIntegrationTable(cmd, integration)
}
