package ksql

import (
	"context"
	"fmt"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/dghubble/sling"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

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
	Id                string `json:"id,omitempty"`
	Name              string `json:"name,omitempty"`
	OutputTopicPrefix string `json:"output_topic_prefix,omitempty"`
	KafkaClusterId    string `json:"kafka_cluster_id,omitempty"`
	Storage           int32  `json:"storage,omitempty"`
	Endpoint          string `json:"endpoint,omitempty"`
	Status            string `json:"status,omitempty"`
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "ksql",
		Short:       "Manage ksqlDB clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &command{
		CLICommand: pcmd.NewCLICommand(cmd, prerunner),
	}

	appCmd := newAppCommand(prerunner)
	clusterCmd := newClusterCommand(cfg, prerunner)

	c.AddCommand(appCmd.Command)
	c.AddCommand(clusterCmd.Command)

	return c.Command
}

// Some helper functions for the ksql app/cluster commands

func (c *ksqlCommand) updateKsqlClusterStatus(cluster *schedv1.KSQLCluster) *ksqlCluster {
	status := cluster.Status.String()
	if cluster.IsPaused {
		status = "PAUSED"
	} else if status == "UP" {
		provisioningFailed, err := c.checkProvisioningFailed(cluster)
		if err != nil {
			status = "UNKNOWN"
		} else if provisioningFailed {
			status = "PROVISIONING FAILED"
		}
	}

	return &ksqlCluster{
		Id:                cluster.Id,
		Name:              cluster.Name,
		OutputTopicPrefix: cluster.OutputTopicPrefix,
		KafkaClusterId:    cluster.KafkaClusterId,
		Storage:           cluster.Storage,
		Endpoint:          cluster.Endpoint,
		Status:            status,
	}
}

func (c *ksqlCommand) checkProvisioningFailed(cluster *schedv1.KSQLCluster) (bool, error) {
	ctx := c.Config.Context()
	state, err := ctx.AuthenticatedState()
	if err != nil {
		return false, err
	}
	bearerToken, err := pauth.GetBearerToken(state, ctx.Platform.Server)
	if err != nil {
		return false, err
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: bearerToken})

	slingClient := sling.New().Client(oauth2.NewClient(context.Background(), ts)).Base(cluster.Endpoint)
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

	return autocompleteClusters(c.EnvironmentId(), c.Client)
}

func autocompleteClusters(environment string, client *ccloud.Client) []string {
	req := &schedv1.KSQLCluster{AccountId: environment}
	clusters, err := client.KSQL.List(context.Background(), req)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(clusters))
	for i, cluster := range clusters {
		suggestions[i] = fmt.Sprintf("%s\t%s", cluster.Id, cluster.Name)
	}
	return suggestions
}
