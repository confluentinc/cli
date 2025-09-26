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

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
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
				Text: "Describe provider integration \"pi-123456\" in the current environment.",
				Code: "confluent provider-integration v2 describe pi-123456",
			},
			examples.Example{
				Text: "Describe provider integration \"pi-123456\" in environment \"env-789012\".",
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

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	integration, _, err := c.V2Client.ProviderIntegrationV2Client.IntegrationsPimV2Api.GetPimV2Integration(c.V2ApiContext(cmd.Context()), integrationId).Environment(environmentId).Execute()
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)

	out := &providerIntegrationDetailedOut{
		Id:          integration.GetId(),
		DisplayName: integration.GetDisplayName(),
		Provider:    integration.GetProvider(),
		Environment: integration.Environment.GetId(),
		Status:      integration.GetStatus(),
		Usages:      integration.GetUsages(),
	}

	// Add provider-specific configuration details
	if integration.Config != nil {
		if integration.Config.PimV2AzureIntegrationConfig != nil {
			azureConfig := integration.Config.PimV2AzureIntegrationConfig
			out.AzureConfig = &azureConfigOut{
				CustomerTenantId:          azureConfig.GetCustomerAzureTenantId(),
				ConfluentMultiTenantAppId: azureConfig.GetConfluentMultiTenantAppId(),
			}
		}
		if integration.Config.PimV2GcpIntegrationConfig != nil {
			gcpConfig := integration.Config.PimV2GcpIntegrationConfig
			out.GcpConfig = &gcpConfigOut{
				CustomerServiceAccount: gcpConfig.GetCustomerGoogleServiceAccount(),
				GoogleServiceAccount:   gcpConfig.GetGoogleServiceAccount(),
			}
		}
	}

	table.Add(out)

	// Filter fields based on provider and configuration state
	fields := []string{"Id", "DisplayName", "Provider", "Environment", "Status"}
	if len(integration.GetUsages()) > 0 {
		fields = append(fields, "Usages")
	}

	switch integration.GetProvider() {
	case providerAzure:
		if out.AzureConfig != nil {
			fields = append(fields, "AzureConfig")
		}
	case providerGcp:
		if out.GcpConfig != nil {
			fields = append(fields, "GcpConfig")
		}
	}

	table.Filter(fields)
	return table.Print()
}

type providerIntegrationDetailedOut struct {
	Id          string          `human:"ID" serialized:"id"`
	DisplayName string          `human:"Name" serialized:"display_name"`
	Provider    string          `human:"Provider" serialized:"provider"`
	Environment string          `human:"Environment" serialized:"environment"`
	Status      string          `human:"Status" serialized:"status"`
	Usages      []string        `human:"Usages" serialized:"usages"`
	AzureConfig *azureConfigOut `human:"Azure Configuration" serialized:"azure_config,omitempty"`
	GcpConfig   *gcpConfigOut   `human:"GCP Configuration" serialized:"gcp_config,omitempty"`
}

type azureConfigOut struct {
	CustomerTenantId          string `human:"Customer Azure Tenant ID" serialized:"customer_azure_tenant_id"`
	ConfluentMultiTenantAppId string `human:"Confluent Multi-Tenant App ID" serialized:"confluent_multi_tenant_app_id"`
}

func (a *azureConfigOut) String() string {
	return fmt.Sprintf("Customer Azure Tenant ID: %s\nConfluent Multi-Tenant App ID: %s", a.CustomerTenantId, a.ConfluentMultiTenantAppId)
}

type gcpConfigOut struct {
	CustomerServiceAccount string `human:"Customer Google Service Account" serialized:"customer_google_service_account"`
	GoogleServiceAccount   string `human:"Google Service Account" serialized:"google_service_account"`
}

func (g *gcpConfigOut) String() string {
	return fmt.Sprintf("Customer Google Service Account: %s\nGoogle Service Account: %s", g.CustomerServiceAccount, g.GoogleServiceAccount)
}
