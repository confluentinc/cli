package quotas

import (
	"context"
	"fmt"

	quotasv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/quotas/v2"


	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"

)

type command struct {
	*pcmd.AuthenticatedCLICommand
	analyticsClient     analytics.Client
}

type quotaLimit struct {
	Id       string
	DisplayName string
	Scope string
	AppliedLimit int32
	OrganizationId string
	EnvironmentId    string
	KafkaClusterId string
	NetworkId string
	UserId string
}

var (
	listFields             = []string{"QuotaCode", "DisplayName", "Scope", "AppliedLimit", "Organization", "Environment", "Network", "KafkaClusterId", "User"}
	listHumanLabels        = []string{"QuotaCode", "DisplayName", "Scope", "AppliedLimit","Organization", "Environment", "Network", "KafkaClusterId", "User"}
	listStructuredLabels   = []string{"QuotaCode", "DisplayName", "Scope", "AppliedLimit","Organization", "Environment", "Network", "KafkaClusterId", "User"}
)

// New returns the Cobra command for `environment`.
func New(cliName string, prerunner pcmd.PreRunner, analyticsClient analytics.Client) *command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "quota-limits",
			Short: fmt.Sprintf("Look up %s service quotas limits", cliName),
		}, prerunner)
	cmd := &command{AuthenticatedCLICommand: cliCmd, analyticsClient: analyticsClient}
	cmd.init()
	return cmd
}

func (c *command) init() {
	listCmd := &cobra.Command{
		Use:   "list <quota-scope> [--quotacode <quota-code> --kafkacluster <kafkacluster-id> --environment <environment-id> --network <network-id>]",
		Short: "List Confluent Cloud service quota limits by quota scopes. (organization, environment, network, kafka_cluster, service_account or user_account)",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.list),
	}
	listCmd.Flags().StringP("quotacode", "Q", "", "filter the result by quota code")
	listCmd.Flags().StringP("kafkacluster", "K", "", "filter the result by kafka cluster id")
	listCmd.Flags().StringP("environment", "E", "", "filter the result by environment id")
	listCmd.Flags().StringP("network", "N", "", "filter the result by network id")
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)
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

	req := c.QuotasClient.AppliedQuotaQuotasV2Api.ListQuotasV2AppliedQuota(c.createContext()).
		Scope(quotaScope)
	ls, _, err := req.Execute()
	if err != nil {
		return err
	}
	qtls := ls.Data

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
			Id: *qt.Id,
			DisplayName: *qt.DisplayName,
			Scope: *qt.Scope,
			AppliedLimit: *qt.AppliedLimit,
		}
		if qt.Organization != nil {
			outQt.OrganizationId = qt.Organization.Id
		}
		if qt.Environment != nil {
			outQt.EnvironmentId = qt.Environment.Id
		}
		if qt.Network != nil {
			outQt.NetworkId = qt.Network.Id
		}
		if qt.KafkaCluster != nil {
			outQt.KafkaClusterId = qt.KafkaCluster.Id
		}
		if qt.User != nil {
			outQt.UserId = qt.User.Id
		}
		outputWriter.AddElement(outQt)
	}

	return outputWriter.Out()
}

func (c *command) Cmd() *cobra.Command {
	return c.Command
}
