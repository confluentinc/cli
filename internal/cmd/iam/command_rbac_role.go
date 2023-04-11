package iam

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

type roleCommand struct {
	*pcmd.AuthenticatedCLICommand
	cfg *v1.Config
}

type prettyRole struct {
	Name         string `human:"Name"`
	AccessPolicy string `human:"Access Policy"`
}

func newRoleCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "role",
		Short: "Manage RBAC and IAM roles.",
		Long:  "Manage Role-Based Access Control (RBAC) and Identity and Access Management (IAM) roles.",
	}

	c := &roleCommand{cfg: cfg}

	if cfg.IsOnPremLogin() {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)
	} else {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)
	}

	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())

	return cmd
}

func (c *roleCommand) createContext() context.Context {
	if c.cfg.IsCloudLogin() {
		return context.WithValue(context.Background(), mdsv2alpha1.ContextAccessToken, c.Context.GetAuthToken())
	} else {
		return context.WithValue(context.Background(), mds.ContextAccessToken, c.Context.GetAuthToken())
	}
}

func createPrettyRole(role mds.Role) (*prettyRole, error) {
	marshalled, err := json.Marshal(role.AccessPolicy)
	if err != nil {
		return nil, err
	}

	return &prettyRole{
		role.Name,
		string(pretty.Pretty(marshalled)),
	}, nil
}

func createPrettyRoleV2(role mdsv2alpha1.Role) (*prettyRole, error) {
	marshalled, err := json.Marshal(role.Policies)
	if err != nil {
		return nil, err
	}

	return &prettyRole{
		role.Name,
		string(pretty.Pretty(marshalled)),
	}, nil
}
