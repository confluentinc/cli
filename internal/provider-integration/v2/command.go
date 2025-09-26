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
	"context"

	"github.com/spf13/cobra"

	piv2 "github.com/confluentinc/ccloud-sdk-go-v2/provider-integration/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func (c *command) V2ApiContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, piv2.ContextAccessToken, c.Context.GetAuthToken())
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "v2",
		Short:       "Manage provider integrations (v2).",
		Long:        "Manage provider integrations that allow users to configure access to cloud provider resources through Confluent resources.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &command{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
	}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())

	return cmd
}

const (
	// Provider constants
	providerAzure = "azure"
	providerGcp   = "gcp"

	// Status constants
	statusDraft    = "DRAFT"
	statusCreated  = "CREATED"
	statusActive   = "ACTIVE"
	statusInactive = "INACTIVE"
)