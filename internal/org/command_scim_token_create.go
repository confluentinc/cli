package org

import (
	"github.com/spf13/cobra"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
)

func (c *scimTokenCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an org scim token.",
		Args:  cobra.NoArgs,
		RunE:  c.create,
	}

	// Required flags

	// Optional flags

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *scimTokenCommand) create(cmd *cobra.Command, args []string) error {
	createReq := orgv2.InlineObject{}

	scimToken, httpResp, err := c.V2Client.CreateOrgScimToken(createReq)
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}

	return printScimToken(cmd, scimToken)
}
