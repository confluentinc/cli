package endpoint

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type out struct {
	Id             string `human:"ID" serialized:"id"`
	Cloud          string `human:"Cloud" serialized:"cloud"`
	Region         string `human:"Region" serialized:"region"`
	Service        string `human:"Service" serialized:"service"`
	IsPrivate      bool   `human:"Is Private" serialized:"is_private"`
	ConnectionType string `human:"Connection Type,omitempty" serialized:"connection_type,omitempty"`
	Endpoint       string `human:"Endpoint URL" serialized:"endpoint"`
	EndpointType   string `human:"Endpoint Type,omitempty" serialized:"endpoint_type,omitempty"`
	Environment    string `human:"Environment" serialized:"environment"`
	Resource       string `human:"Resource,omitempty" serialized:"resource,omitempty"`
	Gateway        string `human:"Gateway,omitempty" serialized:"gateway,omitempty"`
	AccessPoint    string `human:"Access Point,omitempty" serialized:"access_point,omitempty"`
}

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "endpoint",
		Short:       "Manage Confluent Cloud endpoints.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newListCommand())

	return cmd
}
