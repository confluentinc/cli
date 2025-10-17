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

	piv2 "github.com/confluentinc/ccloud-sdk-go-v2/provider-integration/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <display-name>",
		Short: "Create a provider integration.",
		Long:  "Create a provider integration that allows users to manage access to cloud provider resources through Confluent resources. The integration is created in DRAFT state.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create Azure provider integration "azure-integration" in the current environment.`,
				Code: "confluent provider-integration v2 create azure-integration --cloud azure",
			},
			examples.Example{
				Text: `Create GCP provider integration "gcp-integration" in environment "env-123456".`,
				Code: "confluent provider-integration v2 create gcp-integration --cloud gcp --environment env-123456",
			},
		),
	}

	cmd.Flags().String("cloud", "", fmt.Sprintf("Cloud provider (%s or %s).", providerAzure, providerGcp))
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

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	// Create the integration in DRAFT state
	request := piv2.PimV2Integration{
		DisplayName: &displayName,
		Provider:    &cloud,
		Environment: &piv2.ObjectReference{Id: environmentId},
	}

	providerIntegration, err := c.V2Client.CreatePimV2Integration(cmd.Context(), request)
	if err != nil {
		return err
	}

	// Display the created integration with Confluent-managed identity
	table := output.NewTable(cmd)
	out := &providerIntegrationOut{
		Id:          providerIntegration.GetId(),
		DisplayName: providerIntegration.GetDisplayName(),
		Provider:    providerIntegration.GetProvider(),
		Environment: providerIntegration.Environment.GetId(),
	}

	// Add the Confluent-managed service account/app ID
	if providerIntegration.Config != nil {
		if providerIntegration.Config.PimV2AzureIntegrationConfig != nil {
			azureConfig := providerIntegration.Config.PimV2AzureIntegrationConfig
			out.ConfluentMultiTenantAppId = azureConfig.GetConfluentMultiTenantAppId()
		}
		if providerIntegration.Config.PimV2GcpIntegrationConfig != nil {
			gcpConfig := providerIntegration.Config.PimV2GcpIntegrationConfig
			out.GoogleServiceAccount = gcpConfig.GetGoogleServiceAccount()
		}
	}

	table.Add(out)

	// Filter fields based on provider to avoid showing empty/irrelevant fields
	fields := []string{"Id", "DisplayName", "Provider", "Environment"}
	switch cloud {
	case providerAzure:
		fields = append(fields, "ConfluentMultiTenantAppId")
	case providerGcp:
		fields = append(fields, "GoogleServiceAccount")
	}

	table.Filter(fields)
	return table.Print()
}

