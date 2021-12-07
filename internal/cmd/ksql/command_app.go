package ksql

import (
	"context"
	"fmt"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/dghubble/sling"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type appCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	completableChildren     []*cobra.Command
	completableFlagChildren map[string][]*cobra.Command
	analyticsClient         analytics.Client
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

func newAppCommand(prerunner pcmd.PreRunner, analyticsClient analytics.Client) *appCommand {
	cmd := &cobra.Command{
		Use:         "app",
		Short:       "Manage ksqlDB apps.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &appCommand{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner, subcommandFlags),
		analyticsClient:               analyticsClient,
	}

	createCmd := c.newCreateCommand()
	describeCmd := c.newDescribeCommand()
	deleteCmd := c.newDeleteCommand()
	configureAclsCmd := c.newConfigureAclsCommand()

	c.AddCommand(c.newListCommand())
	c.AddCommand(createCmd)
	c.AddCommand(describeCmd)
	c.AddCommand(deleteCmd)
	c.AddCommand(configureAclsCmd)

	c.completableChildren = []*cobra.Command{describeCmd, deleteCmd, configureAclsCmd}
	c.completableFlagChildren = map[string][]*cobra.Command{"cluster": {createCmd}}

	return c
}

func (c *appCommand) updateKsqlClusterStatus(cluster *schedv1.KSQLCluster) *ksqlCluster {
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

func (c *appCommand) checkProvisioningFailed(cluster *schedv1.KSQLCluster) (bool, error) {
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
