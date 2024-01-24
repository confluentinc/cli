package network

import (
	"strings"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newPrivateLinkAccessCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a private link access.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.privateLinkAccessCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create an AWS PrivateLink access.",
				Code: "confluent network private-link access create --network n-123456 --cloud aws --cloud-account 123456789012",
			},
			examples.Example{
				Text: "Create a named AWS PrivateLink access.",
				Code: "confluent network private-link access create aws-private-link-access --network n-123456 --cloud aws --cloud-account 123456789012",
			},
			examples.Example{
				Text: "Create a named GCP Private Service Connect access.",
				Code: "confluent network private-link access create gcp-private-link-access --network n-123456 --cloud gcp --cloud-account temp-123456",
			},
			examples.Example{
				Text: "Create a named Azure Private Link access.",
				Code: "confluent network private-link access create azure-private-link-access --network n-123456 --cloud azure --cloud-account 1234abcd-12ab-34cd-1234-123456abcdef",
			},
		),
	}

	addNetworkFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("cloud-account", "", "AWS account ID for the account containing the VPCs you want to connect from using AWS PrivateLink. GCP project ID for the account containing the VPCs that you want to connect from using Private Service Connect. Azure subscription ID for the account containing the VNets you want to connect from using Azure Private Link.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("network"))
	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("cloud-account"))

	return cmd
}

func (c *command) privateLinkAccessCreate(cmd *cobra.Command, args []string) error {
	name := ""
	if len(args) == 1 {
		name = args[0]
	}

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}
	cloud = strings.ToUpper(cloud)

	network, err := cmd.Flags().GetString("network")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cloudAccount, err := cmd.Flags().GetString("cloud-account")
	if err != nil {
		return err
	}

	createPrivateLinkAccess := networkingv1.NetworkingV1PrivateLinkAccess{
		Spec: &networkingv1.NetworkingV1PrivateLinkAccessSpec{
			Environment: &networkingv1.ObjectReference{Id: environmentId},
			Network:     &networkingv1.ObjectReference{Id: network},
		},
	}

	if name != "" {
		createPrivateLinkAccess.Spec.SetDisplayName(name)
	}

	switch cloud {
	case CloudAws:
		createPrivateLinkAccess.Spec.Cloud = &networkingv1.NetworkingV1PrivateLinkAccessSpecCloudOneOf{
			NetworkingV1AwsPrivateLinkAccess: &networkingv1.NetworkingV1AwsPrivateLinkAccess{
				Kind:    "AwsPrivateLinkAccess",
				Account: cloudAccount,
			},
		}
	case CloudAzure:
		createPrivateLinkAccess.Spec.Cloud = &networkingv1.NetworkingV1PrivateLinkAccessSpecCloudOneOf{
			NetworkingV1AzurePrivateLinkAccess: &networkingv1.NetworkingV1AzurePrivateLinkAccess{
				Kind:         "AzurePrivateLinkAccess",
				Subscription: cloudAccount,
			},
		}
	case CloudGcp:
		createPrivateLinkAccess.Spec.Cloud = &networkingv1.NetworkingV1PrivateLinkAccessSpecCloudOneOf{
			NetworkingV1GcpPrivateServiceConnectAccess: &networkingv1.NetworkingV1GcpPrivateServiceConnectAccess{
				Kind:    "GcpPrivateServiceConnectAccess",
				Project: cloudAccount,
			},
		}
	}

	access, err := c.V2Client.CreatePrivateLinkAccess(createPrivateLinkAccess)
	if err != nil {
		return err
	}

	return printPrivateLinkAccessTable(cmd, access)
}
