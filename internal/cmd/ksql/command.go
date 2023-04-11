package ksql

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	ksqlv2 "github.com/confluentinc/ccloud-sdk-go-v2/ksql/v2"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

type ksqlCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

type ksqlCluster struct {
	Id                    string `human:"ID" serialized:"id"`
	Name                  string `human:"Name" serialized:"name"`
	OutputTopicPrefix     string `human:"Topic Prefix" serialized:"topic_prefix"`
	KafkaClusterId        string `human:"Kafka" serialized:"kafka"`
	Storage               int32  `human:"Storage" serialized:"storage"`
	Endpoint              string `human:"Endpoint" serialized:"endpoint"`
	Status                string `human:"Status" serialized:"status"`
	DetailedProcessingLog bool   `human:"Detailed Processing Log" serialized:"detailed_processing_log"`
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "ksql",
		Short:       "Manage ksqlDB.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(newClusterCommand(cfg, prerunner))

	return cmd
}

func (c *ksqlCommand) formatClusterForDisplayAndList(cluster *ksqlv2.KsqldbcmV2Cluster) *ksqlCluster {
	detailedProcessingLog := true
	if cluster.Spec.HasUseDetailedProcessingLog() {
		detailedProcessingLog = cluster.Spec.GetUseDetailedProcessingLog()
	}

	return &ksqlCluster{
		Id:                    cluster.GetId(),
		Name:                  cluster.Spec.GetDisplayName(),
		OutputTopicPrefix:     cluster.Status.GetTopicPrefix(),
		KafkaClusterId:        cluster.Spec.KafkaCluster.GetId(),
		Storage:               cluster.Status.GetStorage(),
		Endpoint:              cluster.Status.GetHttpEndpoint(),
		Status:                c.getClusterStatus(cluster),
		DetailedProcessingLog: detailedProcessingLog,
	}
}

func (c *ksqlCommand) getClusterStatus(cluster *ksqlv2.KsqldbcmV2Cluster) string {
	status := cluster.Status.GetPhase()
	if cluster.Status.GetIsPaused() {
		status = "PAUSED"
	} else if status == "PROVISIONED" {
		provisioningFailed, err := c.checkProvisioningFailed(cluster.GetId(), cluster.Status.GetHttpEndpoint())
		if err != nil {
			status = "UNKNOWN"
		} else if provisioningFailed {
			status = "PROVISIONING FAILED"
		}
	}
	return status
}

func (c *ksqlCommand) checkProvisioningFailed(clusterId, endpoint string) (bool, error) {
	ctx := c.Config.Context()
	state, err := ctx.AuthenticatedState()
	if err != nil {
		return false, err
	}
	bearerToken, err := pauth.GetBearerToken(state, ctx.Platform.Server, clusterId)
	if err != nil {
		return false, err
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: bearerToken})

	slingClient := sling.New().Client(oauth2.NewClient(context.Background(), ts)).Base(endpoint)
	var failure map[string]any
	response, err := slingClient.New().Get("/info").Receive(nil, &failure)
	if err != nil || response == nil {
		return false, err
	}

	if response.StatusCode == http.StatusServiceUnavailable {
		errorCode, ok := failure["error_code"].(float64)
		if !ok {
			return false, fmt.Errorf("failed to cast 'error_code' to float64")
		}
		// If we have a 50321 we know that ACLs are misconfigured
		if int(errorCode) == 50321 {
			return true, nil
		}
	}
	return false, nil
}

func (c *ksqlCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	return autocompleteClusters(environmentId, c.V2Client)
}

func autocompleteClusters(environment string, client *ccloudv2.Client) []string {
	clusters, err := client.ListKsqlClusters(environment)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(clusters))
	for i, cluster := range clusters {
		suggestions[i] = fmt.Sprintf("%s\t%s", cluster.GetId(), cluster.Spec.GetDisplayName())
	}
	return suggestions
}
