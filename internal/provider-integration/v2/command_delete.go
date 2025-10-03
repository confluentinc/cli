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
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete one or more provider integrations.",
		Long:  "Delete one or more provider integrations. This operation cannot be undone.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete provider integration \"pi-123456\" in the current environment.",
				Code: "confluent provider-integration v2 delete pi-123456",
			},
			examples.Example{
				Text: "Delete provider integrations \"pi-123456\" and \"pi-789012\" in environment \"env-345678\".",
				Code: "confluent provider-integration v2 delete pi-123456 pi-789012 --environment env-345678",
			},
		),
	}

	cmd.Flags().Bool("force", false, "Skip the deletion confirmation prompt.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		integration, err := c.V2Client.GetPimV2Integration(cmd.Context(), id, environmentId)
		if err != nil {
			return false
		}
		// Check if the integration has any usages
		if len(integration.GetUsages()) > 0 {
			cmd.Printf("Warning: provider integration %q is currently in use by %d resource(s). Remove all usages before deleting.\n", id, len(integration.GetUsages()))
			return false
		}
		return true
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, "provider integration"); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeletePimV2Integration(cmd.Context(), id, environmentId)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, "provider integration")
	return err
}
