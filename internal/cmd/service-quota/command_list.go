package servicequota

import (
	"strconv"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type quotaValue struct {
	QuotaCode    string `human:"Quota Code" serialized:"quota_code"`
	DisplayName  string `human:"Name" serialized:"name"`
	Scope        string `human:"Scope" serialized:"scope"`
	AppliedLimit int32  `human:"Applied Limit" serialized:"applied_limit"`
	// The usage field is actually an integer, but this field is not a required one.
	// Set to an empty string if it does not exist.
	Usage        string `human:"Usage" serialized:"usage"`
	Organization string `human:"Organization" serialized:"organization"`
	Environment  string `human:"Environment" serialized:"environment"`
	KafkaCluster string `human:"Kafka Cluster" serialized:"kafka_cluster"`
	Network      string `human:"Network" serialized:"network"`
	User         string `human:"User" serialized:"user"`
}

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <scope>",
		Short: "List Confluent Cloud service quota values by a scope.",
		Long:  "List Confluent Cloud service quota values by a scope (organization, environment, network, kafka_cluster, service_account, or user_account).",
		Args:  cobra.ExactArgs(1),
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List Confluent Cloud service quota values for scope "organization".`,
				Code: "confluent service-quota list organization",
			},
		),
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("quota-code", "", "Filter the result by quota code.")
	cmd.Flags().String("network", "", "Filter the result by network id.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, args []string) error {
	quotaScope := args[0]

	cluster, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return err
	}
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	network, err := cmd.Flags().GetString("network")
	if err != nil {
		return err
	}

	quotaCode, err := cmd.Flags().GetString("quota-code")
	if err != nil {
		return err
	}

	quotas, err := c.V2Client.ListServiceQuotas(quotaScope, cluster, environment, network, quotaCode)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, quota := range quotas {
		out := &quotaValue{
			QuotaCode:    quota.GetId(),
			DisplayName:  quota.GetDisplayName(),
			Scope:        quota.GetScope(),
			AppliedLimit: quota.GetAppliedLimit(),
		}
		if quota.Usage != nil {
			out.Usage = strconv.Itoa(int(quota.GetUsage()))
		}
		if quota.Organization != nil {
			out.Organization = quota.Organization.GetId()
		}
		if quota.Environment != nil {
			out.Environment = quota.Environment.GetId()
		}
		if quota.Network != nil {
			out.Network = quota.Network.GetId()
		}
		if quota.KafkaCluster != nil {
			out.KafkaCluster = quota.KafkaCluster.GetId()
		}
		if quota.User != nil {
			out.User = quota.User.GetId()
		}
		list.Add(out)
	}
	return list.Print()
}
