package ksql

import (
	"context"
	"fmt"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"

	ksql "github.com/confluentinc/ccloud-sdk-go-v2-internal/ksql/v2"
	"github.com/dghubble/sling"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.CLICommand
}

type ksqlCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

// Contains all the fields for listing + describing from the &schedv1.KSQLCluster object
// in scheduler but changes Status to a string so we can have a `PAUSED` option
type ksqlCluster struct {
	Id                    string `json:"id,omitempty"`
	Name                  string `json:"name,omitempty"`
	OutputTopicPrefix     string `json:"output_topic_prefix,omitempty"`
	KafkaClusterId        string `json:"kafka_cluster_id,omitempty"`
	Storage               int32  `json:"storage,omitempty"`
	Endpoint              string `json:"endpoint,omitempty"`
	Status                string `json:"status,omitempty"`
	DetailedProcessingLog bool   `json:"detailed_processing_log,omitempty"`
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "ksql",
		Short:       "Manage ksqlDB.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &command{pcmd.NewCLICommand(cmd, prerunner)}

	c.AddCommand(newAppCommand(prerunner))
	c.AddCommand(newClusterCommand(cfg, prerunner))

	return c.Command
}

// Some helper functions for the ksql app/cluster commands

func (c *ksqlCommand) convertV1ToSchedV2Subset(cluster *schedv1.KSQLCluster) *ksql.KsqldbcmV2Cluster {
	return &ksql.KsqldbcmV2Cluster{
		Spec: &ksql.KsqldbcmV2ClusterSpec{
			DisplayName: &cluster.Name,
			KafkaCluster: &ksql.ObjectReference{
				Id:          cluster.KafkaClusterId,
				Environment: &cluster.AccountId,
			},
			Environment: &ksql.ObjectReference{
				Id: cluster.AccountId,
			},
		},
		Status: &ksql.KsqldbcmV2ClusterStatus{
			HttpEndpoint: &cluster.Endpoint,
			Phase:        cluster.Status.String(),
			TopicPrefix:  &cluster.OutputTopicPrefix,
		},
	}
}

func (c *ksqlCommand) formatClusterForDisplayAndList(cluster *ksql.KsqldbcmV2Cluster) *ksqlCluster {
	status := cluster.Status.Phase
	if cluster.Status.IsPaused {
		status = "PAUSED"
	} else if status == "PROVISIONED" {
		provisioningFailed, err := c.checkProvisioningFailed(*cluster.Id, cluster.Status.GetHttpEndpoint())
		if err != nil {
			status = "UNKNOWN"
		} else if provisioningFailed {
			status = "PROVISIONING FAILED"
		}
	}

	detailedProcessingLog := true
	if cluster.Spec.UseDetailedProcessingLog != nil {
		detailedProcessingLog = *cluster.Spec.UseDetailedProcessingLog
	}

	return &ksqlCluster{
		Id:                    cluster.GetId(),
		Name:                  cluster.Spec.GetDisplayName(),
		OutputTopicPrefix:     cluster.Status.GetTopicPrefix(),
		KafkaClusterId:        cluster.Spec.KafkaCluster.GetId(),
		Storage:               cluster.Status.Storage,
		Endpoint:              cluster.Status.GetHttpEndpoint(),
		Status:                status,
		DetailedProcessingLog: detailedProcessingLog,
	}
}

// checkProvisioningFailed checks if ACLs are misconfigured on the
// cluster.
//
// Send a GET request to the cluster's /info endpoint using oauth
// token from context. If the response contains status code 503 and a
// 50321 error_code, return (true, nil)
// Otherwise, return (false, err (or nil))
//
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
	var failure map[string]interface{}
	response, err := slingClient.New().Get("/info").Receive(nil, &failure)
	if err != nil || response == nil {
		return false, err
	}

	if response.StatusCode == 503 {
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

	return autocompleteClusters(c.EnvironmentId(), c.V2Client)
}

func autocompleteClusters(environment string, client *ccloudv2.Client) []string {
	clusters, err := client.ListKsqlClusters(environment)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(clusters.Data))
	for i, cluster := range clusters.Data {
		suggestions[i] = fmt.Sprintf("%s\t%s", *cluster.Id, *cluster.Spec.DisplayName)
	}
	return suggestions
}
