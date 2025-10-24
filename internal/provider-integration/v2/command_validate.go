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

	piv2 "github.com/confluentinc/ccloud-sdk-go-v2/provider-integration/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <id>",
		Short: "Validate a provider integration.",
		Long:  "Validate that a provider integration is correctly configured with the cloud provider.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.validate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Validate Azure provider integration "cspi-123456".`,
				Code: "confluent provider-integration v2 validate cspi-123456",
			},
			examples.Example{
				Text: `Validate GCP provider integration "cspi-789012".`,
				Code: "confluent provider-integration v2 validate cspi-789012",
			},
		),
	}

	cmd.Flags().String("azure-tenant-id", "", "Customer Azure Tenant ID (for validating Azure provider before update).")
	cmd.Flags().String("gcp-service-account", "", "Customer Google Service Account (for validating GCP provider before update).")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) validate(cmd *cobra.Command, args []string) error {
	integrationId := args[0]

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	// Get optional config for validation before update
	azureTenantId, _ := cmd.Flags().GetString("azure-tenant-id")
	gcpServiceAccount, _ := cmd.Flags().GetString("gcp-service-account")

	validateReq := piv2.PimV2IntegrationValidateRequest{
		Id:          &integrationId,
		Environment: &piv2.GlobalObjectReference{Id: environmentId},
	}

	// Add config if provided (allows validation before update)
	if azureTenantId != "" || gcpServiceAccount != "" {
		var validateConfig piv2.PimV2IntegrationValidateRequestConfigOneOf

		if azureTenantId != "" {
			azureConfig := &piv2.PimV2AzureIntegrationConfig{
				Kind:                  "AzureIntegrationConfig",
				CustomerAzureTenantId: &azureTenantId,
			}
			validateConfig = piv2.PimV2AzureIntegrationConfigAsPimV2IntegrationValidateRequestConfigOneOf(azureConfig)
		} else if gcpServiceAccount != "" {
			gcpConfig := &piv2.PimV2GcpIntegrationConfig{
				Kind:                         "GcpIntegrationConfig",
				CustomerGoogleServiceAccount: &gcpServiceAccount,
			}
			validateConfig = piv2.PimV2GcpIntegrationConfigAsPimV2IntegrationValidateRequestConfigOneOf(gcpConfig)
		}

		validateReq.Config = &validateConfig
	}

	if err := c.V2Client.ValidatePimV2Integration(cmd.Context(), validateReq); err != nil {
		return err
	}

	cmd.Printf("Successfully validated provider integration %q.\n", integrationId)
	return nil
}
