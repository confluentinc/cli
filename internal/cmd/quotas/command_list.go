package quotas

import (
	"context"
	"net/url"

	quotasv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/quotas/v2"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type quotaLimit struct {
	QuotaCode    string
	DisplayName  string
	Scope        string
	AppliedLimit int32
	Organization string
	Environment  string
	KafkaCluster string
	Network      string
	User         string
}

var (
	listFields           = []string{"QuotaCode", "DisplayName", "Scope", "AppliedLimit", "Organization", "Environment", "Network", "KafkaCluster", "User"}
	listHumanLabels      = []string{"Quota Code", "Display Name", "Scope", "Applied Limit", "Organization", "Environment", "Network", "Kafka Cluster", "User"}
	listStructuredLabels = []string{"quota_code", "display_name", "scope", "applied_limit", "organization", "environment", "network", "kafka_cluster", "user"}
)

func (c *command) newListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list <quota-scope> [--quota-code <quota-code> --kafka-cluster <kafka-cluster-id> --environment <environment-id> --network <network-id>]",
		Short: "List Confluent Cloud service quota limits by quota scopes. (organization, environment, network, kafka_cluster, service_account or user_account)",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.list),
	}
	listCmd.Flags().StringP("quota-code", "Q", "", "filter the result by quota code")
	listCmd.Flags().StringP("kafka-cluster", "K", "", "filter the result by kafka cluster id")
	listCmd.Flags().StringP("environment", "E", "", "filter the result by environment id")
	listCmd.Flags().StringP("network", "N", "", "filter the result by network id")
	listCmd.Flags().SortFlags = false
	pcmd.AddOutputFlag(listCmd)
	return listCmd
}

func (c *command) createContext() context.Context {
	return context.WithValue(context.Background(), quotasv2.ContextAccessToken, c.State.AuthToken)
}

func (c *command) list(cmd *cobra.Command, args []string) error {
	quotaScope := args[0]

	quotaCode, err := cmd.Flags().GetString("quota-code")
	if err != nil {
		return err
	}
	kafkaCluster, err := cmd.Flags().GetString("kafka-cluster")
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

	firstTime := true
	token := ""
	qtls := []quotasv2.QuotasV2AppliedQuota{}

	// Since we use paginated results, get all results by iterating the list.
	for token != "" || firstTime {
		firstTime = false
		req := c.QuotasClient.AppliedQuotaQuotasV2Api.ListQuotasV2AppliedQuota(c.createContext()).
			Scope(quotaScope).PageToken(token)
		lsResult, _, err := req.Execute()
		if err != nil {
			return err
		}
		qtls = append(qtls, lsResult.Data...)

		token = ""
		if md, ok := lsResult.GetMetadataOk(); ok && md.GetNext() != "" {
			url, err := url.Parse(*lsResult.GetMetadata().Next.Get())
			if err != nil {
				return err
			}
			token = url.Query().Get("page_token")
		}
	}

	//filter by quota id
	filtered := []quotasv2.QuotasV2AppliedQuota{}
	if quotaCode != "" {
		for _, qt := range qtls {
			if *qt.Id == quotaCode {
				filtered = append(filtered, qt)
			}
		}
		qtls = filtered
	}

	//filter by environment id
	filtered = []quotasv2.QuotasV2AppliedQuota{}
	if environment != "" {
		for _, qt := range qtls {
			if qt.Environment != nil && qt.Environment.Id == environment {
				filtered = append(filtered, qt)
			}
		}
		qtls = filtered
	}

	//filter by cluster id
	filtered = []quotasv2.QuotasV2AppliedQuota{}
	if kafkaCluster != "" {
		for _, qt := range qtls {
			if qt.KafkaCluster != nil && qt.KafkaCluster.Id == kafkaCluster {
				filtered = append(filtered, qt)
			}
		}
		qtls = filtered
	}

	//filter by network id
	filtered = []quotasv2.QuotasV2AppliedQuota{}
	if network != "" {
		for _, qt := range qtls {
			if qt.Network != nil && qt.Network.Id == network {
				filtered = append(filtered, qt)
			}
		}
		qtls = filtered
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}

	for _, qt := range qtls {
		outQt := &quotaLimit{
			QuotaCode:    *qt.Id,
			DisplayName:  *qt.DisplayName,
			Scope:        *qt.Scope,
			AppliedLimit: *qt.AppliedLimit,
		}
		if qt.Organization != nil {
			outQt.Organization = qt.Organization.Id
		}
		if qt.Environment != nil {
			outQt.Environment = qt.Environment.Id
		}
		if qt.Network != nil {
			outQt.Network = qt.Network.Id
		}
		if qt.KafkaCluster != nil {
			outQt.KafkaCluster = qt.KafkaCluster.Id
		}
		if qt.User != nil {
			outQt.User = qt.User.Id
		}
		outputWriter.AddElement(outQt)
	}

	return outputWriter.Out()
}

func (c *command) Cmd() *cobra.Command {
	return c.Command
}
