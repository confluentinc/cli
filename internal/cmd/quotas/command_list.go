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

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <quota-scope>",
		Short: "List Confluent Cloud service quota limits by a scope.",
		Long:  "List Confluent Cloud service quota limits by a scope (organization, environment, network, kafka_cluster, service_account, or user_account).",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.list),
	}
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("quota-code", "", "Filter the result by quota code.")
	cmd.Flags().String("network", "", "Filter the result by network id.")
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

	token := ""
	quotaList := []quotasv2.QuotasV2AppliedQuota{}

	// Since we use paginated results, get all results by iterating the list.
	for {
		req := c.QuotasClient.AppliedQuotaQuotasV2Api.ListQuotasV2AppliedQuota(c.createContext()).
			Scope(quotaScope).PageToken(token)
		lsResult, _, err := req.Execute()
		if err != nil {
			return err
		}
		quotaList = append(quotaList, lsResult.Data...)

		token = ""
		if md, ok := lsResult.GetMetadataOk(); ok && md.GetNext() != "" {
			url, err := url.Parse(*md.Next.Get())
			if err != nil {
				return err
			}
			token = url.Query().Get("page_token")
		}

		if token == "" {
			break
		}
	}

	quotaList = filterQuotaResults(quotaList, quotaCode, environment, network, kafkaCluster)

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}

	for _, quota := range quotaList {
		outQt := &quotaLimit{
			QuotaCode:    *quota.Id,
			DisplayName:  *quota.DisplayName,
			Scope:        *quota.Scope,
			AppliedLimit: *quota.AppliedLimit,
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
