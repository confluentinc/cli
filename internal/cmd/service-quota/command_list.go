package servicequota

import (
	"strconv"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type quotaValue struct {
	QuotaCode    string
	DisplayName  string
	Scope        string
	AppliedLimit int32
	//The usage field is actually an integer, but this field is not a required one.
	//Set to an empty string if it does not exist.
	Usage        string
	Organization string
	Environment  string
	KafkaCluster string
	Network      string
	User         string
}

var (
	listFields           = []string{"QuotaCode", "DisplayName", "Scope", "AppliedLimit", "Usage", "Organization", "Environment", "Network", "KafkaCluster", "User"}
	listHumanLabels      = []string{"Quota Code", "Display Name", "Scope", "Applied Limit", "Usage", "Organization", "Environment", "Network", "Kafka Cluster", "User"}
	listStructuredLabels = []string{"quota_code", "display_name", "scope", "applied_limit", "usage", "organization", "environment", "network", "kafka_cluster", "user"}
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <quota-scope>",
		Short: "List Confluent Cloud service quota values by a scope.",
		Long:  "List Confluent Cloud service quota values by a scope (organization, environment, network, kafka_cluster, service_account, or user_account).",
		Args:  cobra.ExactArgs(1),
		RunE:  c.list,
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

	kafkaCluster, err := cmd.Flags().GetString("cluster")
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

	quotas, err := c.V2Client.ListServiceQuotas(quotaScope, kafkaCluster, environment, network, quotaCode)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}

	for _, quota := range quotas {
		outQt := &quotaValue{
			QuotaCode:    *quota.Id,
			DisplayName:  *quota.DisplayName,
			Scope:        *quota.Scope,
			AppliedLimit: *quota.AppliedLimit,
		}

		if quota.Usage != nil {
			outQt.Usage = strconv.Itoa(int(*quota.Usage))
		}

		if quota.Organization != nil {
			outQt.Organization = quota.Organization.Id
		}
		if quota.Environment != nil {
			outQt.Environment = quota.Environment.Id
		}
		if quota.Network != nil {
			outQt.Network = quota.Network.Id
		}
		if quota.KafkaCluster != nil {
			outQt.KafkaCluster = quota.KafkaCluster.Id
		}
		if quota.User != nil {
			outQt.User = quota.User.Id
		}
		outputWriter.AddElement(outQt)
	}

	return outputWriter.Out()
}
