package iam

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	iamipfilteringv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam-ip-filtering/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *ipFilterCommand) newListCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List IP filters.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}
	if cfg.IsTest || (cfg.Context() != nil && featureflags.Manager.BoolVariation("auth.ip_filter.sr.cli.enabled", cfg.Context(), featureflags.GetCcloudLaunchDarklyClient(cfg.Context().PlatformName), true, false)) {
		cmd.Flags().String("environment", "", "Name of the environment for which this filter applies. By default will apply to the org only.")
		cmd.Flags().Bool("include-parent-scope", true, "If an environment is specified, include organization scoped filters.")
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ipFilterCommand) list(cmd *cobra.Command, _ []string) error {
	var ipFilters []iamipfilteringv2.IamV2IpFilter
	ldClient := featureflags.GetCcloudLaunchDarklyClient(c.Context.PlatformName)
	isSrEnabled := featureflags.Manager.BoolVariation("auth.ip_filter.sr.cli.enabled", c.Context, ldClient, true, false)
	if isSrEnabled {
		orgId := c.Context.GetCurrentOrganization()
		environment, err := cmd.Flags().GetString("environment")
		if err != nil {
			return err
		}
		resourceScope := ""
		if environment != "" {
			resourceScope = fmt.Sprintf(resourceScopeStr, orgId, environment)
		}
		includeParentScope, err := cmd.Flags().GetBool("include-parent-scope")
		if err != nil {
			return err
		}
		parentScope := ""
		if cmd.Flags().Changed("include-parent-scope") {
			parentScope = strconv.FormatBool(includeParentScope)
		}
		ipFilters, err = c.V2Client.ListIamIpFilters(resourceScope, parentScope)
		if err != nil {
			return err
		}
	} else {
		var err error
		ipFilters, err = c.V2Client.ListIamIpFilters("", "")
		if err != nil {
			return err
		}
	}
	list := output.NewList(cmd)
	for _, filter := range ipFilters {
		var filterOut = ipFilterOut{
			ID:            filter.GetId(),
			Name:          filter.GetFilterName(),
			ResourceGroup: filter.GetResourceGroup(),
			IpGroups:      convertIpGroupObjectsToIpGroupIds(filter),
		}
		if isSrEnabled {
			filterOut.ResourceScope = filter.GetResourceScope()
			filterOut.OperationGroups = filter.GetOperationGroups()
		}
		list.Add(filterOut)
	}
	return list.Print()
}

func convertIpGroupObjectsToIpGroupIds(filter iamipfilteringv2.IamV2IpFilter) []string {
	ipGroups := filter.GetIpGroups()
	ipGroupIds := make([]string, len(ipGroups))
	for i, group := range ipGroups {
		ipGroupIds[i] = group.GetId()
	}
	return ipGroupIds
}
