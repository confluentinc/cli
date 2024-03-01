package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

type accessPointCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type accessPointOut struct {
	Id                              string `human:"ID" serialized:"id"`
	Name                            string `human:"Name,omitempty" serialized:"name,omitempty"`
	AwsVpcEndpointService           string `human:"AWS VPC Endpoint Service,omitempty" serialized:"aws_vpc_endpoint_service,omitempty"`
	AzurePrivateLinkPrivateEndpoint string `human:"Azure Private Link Private Endpoint,omitempty" serialized:"azure_private_link_private_endpoint,omitempty"`
	Environment                     string `human:"Environment" serialized:"environment"`
	Gateway                         string `human:"Gateway" serialized:"gateway"`
	Phase                           string `human:"Phase" serialized:"phase"`
}

func newAccessPointCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "access-point",
		Short:       "Manage access points.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &accessPointCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newPrivateLinkCommand())

	return cmd
}
