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
	cmd := &cobra.Command{
		Use:   "list <quota-scope> [--quota-code <quota-code> --kafka-cluster <kafka-cluster-id> --environment <environment-id> --network <network-id>]",
		Short: "List Confluent Cloud service quota limits by a scope.",
		Long:  "List Confluent Cloud service quota limits by a scope (organization, environment, network, kafka_cluster, service_account, or user_account)",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.list),
	}
	cmd.Flags().String("quota-code", "", "filter the result by quota code")
	cmd.Flags().String("kafka-cluster", "", "filter the result by kafka cluster id")
	cmd.Flags().String("environment", "", "filter the result by environment id")
	cmd.Flags().String("network", "", "filter the result by network id")
	cmd.Flags().SortFlags = false
	pcmd.AddOutputFlag(cmd)
	return cmd
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

	qtls = filterQuotaResults(qtls, quotaCode, environment, network, kafkaCluster)

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

func filterQuotaResults(quotaList []quotasv2.QuotasV2AppliedQuota, quotaCode string, environment string, network string, kafkaCluster string) []quotasv2.QuotasV2AppliedQuota {
	//filter by quota id
	filtered := []quotasv2.QuotasV2AppliedQuota{}
	if quotaCode != "" {
		for _, qt := range quotaList {
			if *qt.Id == quotaCode {
				filtered = append(filtered, qt)
			}
		}
		quotaList = filtered
	}

	//filter by environment id
	filtered = []quotasv2.QuotasV2AppliedQuota{}
	if environment != "" {
		for _, qt := range quotaList {
			if qt.Environment != nil && qt.Environment.Id == environment {
				filtered = append(filtered, qt)
			}
		}
		quotaList = filtered
	}

	//filter by cluster id
	filtered = []quotasv2.QuotasV2AppliedQuota{}
	if kafkaCluster != "" {
		for _, qt := range quotaList {
			if qt.KafkaCluster != nil && qt.KafkaCluster.Id == kafkaCluster {
				filtered = append(filtered, qt)
			}
		}
		quotaList = filtered
	}

	//filter by network id
	filtered = []quotasv2.QuotasV2AppliedQuota{}
	if network != "" {
		for _, qt := range quotaList {
			if qt.Network != nil && qt.Network.Id == network {
				filtered = append(filtered, qt)
			}
		}
		quotaList = filtered
	}

	return quotaList
}
