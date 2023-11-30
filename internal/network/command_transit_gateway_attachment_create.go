package network

import (
	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newTransitGatewayAttachmentCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a transit gateway attachment.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.transitGatewayAttachmentCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a transit gateway attachment in AWS.",
				Code: "confluent network transit-gateway-attachment create --network n-123456 --aws-ram-share-arn arn:aws:ram:us-west-2:123456789012:resource-share/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx --aws-transit-gateway tgw-xxxxxxxxxxxxxxxxx --routes 10.0.0.0/16,100.64.0.0/10",
			},
			examples.Example{
				Text: "Create a named transit gateway attachment in AWS.",
				Code: "confluent network transit-gateway-attachment create my-tgw-attachment --network n-123456 --aws-ram-share-arn arn:aws:ram:us-west-2:123456789012:resource-share/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx --aws-transit-gateway tgw-xxxxxxxxxxxxxxxxx --routes 10.0.0.0/16,100.64.0.0/10",
			},
		),
	}

	addNetworkFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("aws-ram-share-arn", "", "AWS Resource Name (ARN) for the AWS Resource Access Manager (RAM) Share of the AWS Transit Gateway that you want Confluent Cloud to be attached to.")
	cmd.Flags().String("aws-transit-gateway", "", "ID of the AWS Transit Gateway that you want Confluent Cloud to be attached to.")
	cmd.Flags().StringSlice("routes", nil, "A comma-separated list of CIDRs.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("network"))
	cobra.CheckErr(cmd.MarkFlagRequired("aws-ram-share-arn"))
	cobra.CheckErr(cmd.MarkFlagRequired("aws-transit-gateway"))
	cobra.CheckErr(cmd.MarkFlagRequired("routes"))

	return cmd
}

func (c *command) transitGatewayAttachmentCreate(cmd *cobra.Command, args []string) error {
	name := ""
	if len(args) == 1 {
		name = args[0]
	}

	network, err := cmd.Flags().GetString("network")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	awsRamShareArn, err := cmd.Flags().GetString("aws-ram-share-arn")
	if err != nil {
		return err
	}

	awsTransitGateway, err := cmd.Flags().GetString("aws-transit-gateway")
	if err != nil {
		return err
	}

	routes, err := cmd.Flags().GetStringSlice("routes")
	if err != nil {
		return err
	}

	createAttachment := networkingv1.NetworkingV1TransitGatewayAttachment{
		Spec: &networkingv1.NetworkingV1TransitGatewayAttachmentSpec{
			Environment: &networkingv1.ObjectReference{Id: environmentId},
			Network:     &networkingv1.ObjectReference{Id: network},
			Cloud: &networkingv1.NetworkingV1TransitGatewayAttachmentSpecCloudOneOf{
				NetworkingV1AwsTransitGatewayAttachment: &networkingv1.NetworkingV1AwsTransitGatewayAttachment{
					Kind:             "AwsTransitGatewayAttachment",
					RamShareArn:      awsRamShareArn,
					Routes:           routes,
					TransitGatewayId: awsTransitGateway,
				},
			},
		},
	}

	if name != "" {
		createAttachment.Spec.SetDisplayName(name)
	}

	attachment, err := c.V2Client.CreateTransitGatewayAttachment(createAttachment)
	if err != nil {
		return err
	}

	return printTransitGatewayAttachmentTable(cmd, attachment)
}
