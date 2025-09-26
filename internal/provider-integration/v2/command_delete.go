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
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a provider integration.",
		Long:  "Delete a provider integration. This operation cannot be undone.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete provider integration \"pi-123456\" in the current environment.",
				Code: "confluent provider-integration v2 delete pi-123456",
			},
			examples.Example{
				Text: "Delete provider integration \"pi-123456\" in environment \"env-789012\".",
				Code: "confluent provider-integration v2 delete pi-123456 --environment env-789012",
			},
		),
	}

	cmd.Flags().Bool("force", false, "Skip the deletion confirmation prompt.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	integrationId := args[0]

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	// First, check if the integration exists and get its status
	integration, _, err := c.V2Client.ProviderIntegrationV2Client.IntegrationsPimV2Api.GetPimV2Integration(c.V2ApiContext(cmd.Context()), integrationId).Environment(environmentId).Execute()
	if err != nil {
		return err
	}

	// Check if the integration has any usages
	if len(integration.GetUsages()) > 0 {
		return fmt.Errorf("cannot delete provider integration %q: it is currently in use by %d resource(s). Remove all usages before deleting", integrationId, len(integration.GetUsages()))
	}

	// Confirm deletion unless --force is used
	promptMsg := deletion.DefaultYesNoPromptString(cmd, "provider integration", []string{integrationId}, "")
	if err := deletion.ConfirmPrompt(cmd, promptMsg); err != nil {
		return err
	}

	if _, err := c.V2Client.ProviderIntegrationV2Client.IntegrationsPimV2Api.DeletePimV2Integration(c.V2ApiContext(cmd.Context()), integrationId).Environment(environmentId).Execute(); err != nil {
		return err
	}

	cmd.Printf("Deleted provider integration \"%s\".\n", integrationId)
	return nil
}
