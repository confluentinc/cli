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
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List provider integrations.",
		Long:  "List all provider integrations in the specified environment.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all provider integrations in the current environment.",
				Code: "confluent provider-integration v2 list",
			},
			examples.Example{
				Text: "List all provider integrations in environment \"env-123456\".",
				Code: "confluent provider-integration v2 list --environment env-123456",
			},
		),
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	integrations, err := c.V2Client.ListPimV2Integrations(cmd.Context(), environmentId)
	if err != nil {
		return err
	}

	list := make([]providerIntegrationListOut, len(integrations))
	for i, integration := range integrations {
		out := providerIntegrationListOut{
			Id:          integration.GetId(),
			DisplayName: integration.GetDisplayName(),
			Provider:    integration.GetProvider(),
			Environment: integration.Environment.GetId(),
			Status:      integration.GetStatus(),
		}

		// Add provider-specific configuration details
		if integration.Config != nil {
			if integration.Config.PimV2AzureIntegrationConfig != nil {
				azureConfig := integration.Config.PimV2AzureIntegrationConfig
				out.CustomerAzureTenantId = azureConfig.GetCustomerAzureTenantId()
				out.ConfluentMultiTenantAppId = azureConfig.GetConfluentMultiTenantAppId()
			}
			if integration.Config.PimV2GcpIntegrationConfig != nil {
				gcpConfig := integration.Config.PimV2GcpIntegrationConfig
				out.CustomerGoogleServiceAccount = gcpConfig.GetCustomerGoogleServiceAccount()
				out.GoogleServiceAccount = gcpConfig.GetGoogleServiceAccount()
			}
		}

		list[i] = out
	}

	outputList := output.NewList(cmd)
	for _, item := range list {
		outputList.Add(&item)
	}

	return outputList.Print()
}
