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

	"github.com/confluentinc/cli/v4/pkg/output"
)

// providerIntegrationOut represents the output structure for provider integrations
type providerIntegrationOut struct {
	Id          string `human:"ID" serialized:"id"`
	DisplayName string `human:"Name" serialized:"display_name"`
	Provider    string `human:"Provider" serialized:"provider"`
	Environment string `human:"Environment" serialized:"environment"`
	Status      string `human:"Status" serialized:"status"`
	// Azure-specific fields (omitempty ensures they're only shown for Azure integrations)
	CustomerAzureTenantId     string `human:"Customer Azure Tenant ID" serialized:"customer_azure_tenant_id,omitempty"`
	ConfluentMultiTenantAppId string `human:"Confluent Azure Multi-Tenant App ID" serialized:"confluent_multi_tenant_app_id,omitempty"`
	// GCP-specific fields (omitempty ensures they're only shown for GCP integrations)
	CustomerGoogleServiceAccount string `human:"Customer Google Service Account" serialized:"customer_google_service_account,omitempty"`
	GoogleServiceAccount         string `human:"Google Service Account" serialized:"google_service_account,omitempty"`
}

func printProviderIntegrationTable(cmd *cobra.Command, integration piv2.PimV2Integration) error {
	out := &providerIntegrationOut{
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

	table := output.NewTable(cmd)
	table.Add(out)
	return table.Print()
}
