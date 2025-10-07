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
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	piv2 "github.com/confluentinc/ccloud-sdk-go-v2/provider-integration/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <display-name>",
		Short: "Create a provider integration.",
		Long:  "Create a provider integration that allows users to manage access to cloud provider resources through Confluent resources.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create and configure Azure provider integration "azure-integration" in the current environment.`,
				Code: "confluent provider-integration v2 create azure-integration --cloud azure --azure-tenant-id 00000000-0000-0000-0000-000000000000",
			},
			examples.Example{
				Text: `Create and configure GCP provider integration "gcp-integration" in environment "env-123456".`,
				Code: "confluent provider-integration v2 create gcp-integration --cloud gcp --gcp-service-account my-sa@my-project.iam.gserviceaccount.com --environment env-123456",
			},
		),
	}

	cmd.Flags().String("cloud", "", fmt.Sprintf("Cloud provider (%s or %s).", providerAzure, providerGcp))
	cmd.Flags().String("azure-tenant-id", "", "Azure Tenant ID (required for Azure provider).")
	cmd.Flags().String("gcp-service-account", "", "Customer Google Service Account (required for GCP provider).")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))

	return cmd
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	displayName := args[0]

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	cloud = strings.ToLower(cloud)
	if cloud != providerAzure && cloud != providerGcp {
		return fmt.Errorf("cloud provider must be either %s or %s", providerAzure, providerGcp)
	}

	// Get provider-specific configuration
	var azureTenantId, gcpServiceAccount string
	switch cloud {
	case providerAzure:
		azureTenantId, err = cmd.Flags().GetString("azure-tenant-id")
		if err != nil {
			return err
		}
		if azureTenantId == "" {
			return fmt.Errorf("--azure-tenant-id is required for Azure provider integrations")
		}
	case providerGcp:
		gcpServiceAccount, err = cmd.Flags().GetString("gcp-service-account")
		if err != nil {
			return err
		}
		if gcpServiceAccount == "" {
			return fmt.Errorf("--gcp-service-account is required for GCP provider integrations")
		}
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	// Step 1: Create the integration in DRAFT state
	request := piv2.PimV2Integration{
		DisplayName: &displayName,
		Provider:    &cloud,
		Environment: &piv2.ObjectReference{Id: environmentId},
	}

	providerIntegration, err := c.V2Client.CreatePimV2Integration(cmd.Context(), request)
	if err != nil {
		return err
	}

	integrationId := providerIntegration.GetId()
	cmd.Printf("Created provider integration %s in DRAFT state.\n", integrationId)

	// Step 2: Authorize with provider-specific configuration
	var updateConfig piv2.PimV2IntegrationUpdateConfigOneOf

	switch cloud {
	case providerAzure:
		azureConfig := &piv2.PimV2AzureIntegrationConfig{
			Kind:                  "AzureIntegrationConfig",
			CustomerAzureTenantId: &azureTenantId,
		}
		updateConfig = piv2.PimV2AzureIntegrationConfigAsPimV2IntegrationUpdateConfigOneOf(azureConfig)
	case providerGcp:
		gcpConfig := &piv2.PimV2GcpIntegrationConfig{
			Kind:                         "GcpIntegrationConfig",
			CustomerGoogleServiceAccount: &gcpServiceAccount,
		}
		updateConfig = piv2.PimV2GcpIntegrationConfigAsPimV2IntegrationUpdateConfigOneOf(gcpConfig)
	}

	updateReq := piv2.PimV2IntegrationUpdate{
		Config:      &updateConfig,
		Environment: &piv2.ObjectReference{Id: environmentId},
	}

	updated, err := c.V2Client.UpdatePimV2Integration(cmd.Context(), integrationId, updateReq)
	if err != nil {
		// Extract the specific error message from the backend response
		errorMsg := err.Error()
		if errorMsg == "" {
			errorMsg = "configuration error"
		}
		
		// If update fails, clean up the draft integration to make the operation atomic
		if deleteErr := c.V2Client.DeletePimV2Integration(cmd.Context(), integrationId, environmentId); deleteErr != nil {
			cmd.Printf("Warning: Failed to clean up draft integration %s: %v\n", integrationId, deleteErr)
		} else {
			cmd.Printf("Deleted draft integration %s due to: %s\n", integrationId, errorMsg)
		}
		return err
	}

	cmd.Printf("Configured %s provider settings.\n", cloud)

	// Step 3: Validate the setup
	validateReq := piv2.PimV2IntegrationValidateRequest{
		Id:          &integrationId,
		Environment: &piv2.GlobalObjectReference{Id: environmentId},
	}

	if err := c.V2Client.ValidatePimV2Integration(cmd.Context(), validateReq); err != nil {
		// Show setup instructions if validation fails
		if updated.Config != nil {
			switch cloud {
			case providerAzure:
				if updated.Config.PimV2AzureIntegrationConfig != nil {
					azureConfig := updated.Config.PimV2AzureIntegrationConfig
					cmd.Println("\n⏳ Azure setup required:")
					cmd.Printf("1. Run: az ad sp create --id %s\n", azureConfig.GetConfluentMultiTenantAppId())
					cmd.Println("2. Check Azure Portal → Enterprise Applications to ensure it appears")
					cmd.Println("3. Grant necessary permissions to the service principal")
					cmd.Printf("\nRun 'confluent provider-integration v2 describe %s' to check status.\n", integrationId)
				}
			case providerGcp:
				if updated.Config.PimV2GcpIntegrationConfig != nil {
					gcpConfig := updated.Config.PimV2GcpIntegrationConfig
					cmd.Println("\n⏳ GCP setup required:")
					cmd.Printf("1. Grant Service Account Token Creator role:\n")
					cmd.Printf("   gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \\\n")
					cmd.Printf("     --member=\"serviceAccount:%s\" \\\n", gcpConfig.GetGoogleServiceAccount())
					cmd.Printf("     --role=\"roles/iam.serviceAccountTokenCreator\" \\\n")
					cmd.Printf("     --condition=\"expression=request.auth.claims.sub=='%s'\"\n\n", gcpConfig.GetGoogleServiceAccount())
					cmd.Printf("2. Grant your service account (%s) permissions based on your connector needs\n", gcpConfig.GetCustomerGoogleServiceAccount())
					cmd.Printf("\nRun 'confluent provider-integration v2 describe %s' to check status.\n", integrationId)
					cmd.Println("\nNote: IAM changes may take 1-7 minutes to propagate.")
				}
			}
		}
	} else {
		cmd.Printf("\n✓ %s setup validated successfully!\n", cases.Title(language.English).String(cloud))
	}

	// Display the final configuration
	table := output.NewTable(cmd)
	out := &providerIntegrationOut{
		Id:          updated.GetId(),
		DisplayName: updated.GetDisplayName(),
		Provider:    updated.GetProvider(),
		Environment: updated.Environment.GetId(),
		Status:      updated.GetStatus(),
	}

	table.Add(out)
	table.Filter([]string{"Id", "DisplayName", "Provider", "Environment", "Status"})
	return table.Print()
}
