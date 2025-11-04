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
	"fmt"

	"github.com/spf13/cobra"

	piv2 "github.com/confluentinc/ccloud-sdk-go-v2/provider-integration/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a provider integration.",
		Long:  "Update an existing provider integration to map customer's identity.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update Azure provider integration "cspi-123456" with customer tenant ID.`,
				Code: "confluent provider-integration v2 update cspi-123456 --azure-tenant-id 00000000-0000-0000-0000-000000000000",
			},
			examples.Example{
				Text: `Update GCP provider integration "cspi-789012" with customer service account.`,
				Code: "confluent provider-integration v2 update cspi-789012 --gcp-service-account my-sa@my-project.iam.gserviceaccount.com",
			},
		),
	}

	cmd.Flags().String("azure-tenant-id", "", "Customer Azure Tenant ID (required for Azure provider).")
	cmd.Flags().String("gcp-service-account", "", "Customer Google Service Account (required for GCP provider).")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
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

	// First, get the integration to determine its provider
	integration, err := c.V2Client.GetPimV2Integration(cmd.Context(), integrationId, environmentId)
	if err != nil {
		return err
	}

	provider := integration.GetProvider()

	// Get provider-specific configuration
	var updateConfig piv2.PimV2IntegrationUpdateConfigOneOf

	switch provider {
	case providerAzure:
		azureTenantId, err := cmd.Flags().GetString("azure-tenant-id")
		if err != nil {
			return err
		}
		if azureTenantId == "" {
			return fmt.Errorf("--azure-tenant-id is required for Azure provider integrations")
		}

		azureConfig := &piv2.PimV2AzureIntegrationConfig{
			Kind:                  "AzureIntegrationConfig",
			CustomerAzureTenantId: &azureTenantId,
		}
		updateConfig = piv2.PimV2AzureIntegrationConfigAsPimV2IntegrationUpdateConfigOneOf(azureConfig)

	case providerGcp:
		gcpServiceAccount, err := cmd.Flags().GetString("gcp-service-account")
		if err != nil {
			return err
		}
		if gcpServiceAccount == "" {
			return fmt.Errorf("--gcp-service-account is required for GCP provider integrations")
		}

		gcpConfig := &piv2.PimV2GcpIntegrationConfig{
			Kind:                         "GcpIntegrationConfig",
			CustomerGoogleServiceAccount: &gcpServiceAccount,
		}
		updateConfig = piv2.PimV2GcpIntegrationConfigAsPimV2IntegrationUpdateConfigOneOf(gcpConfig)

	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	updateReq := piv2.PimV2IntegrationUpdate{
		Config:      &updateConfig,
		Environment: &piv2.ObjectReference{Id: environmentId},
	}

	updated, err := c.V2Client.UpdatePimV2Integration(cmd.Context(), integrationId, updateReq)
	if err != nil {
		return err
	}

	return printProviderIntegrationTable(cmd, updated)
}
