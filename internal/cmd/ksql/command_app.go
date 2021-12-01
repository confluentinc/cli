package ksql

import (
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type appCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner               pcmd.PreRunner
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

func NewClusterCommand(prerunner pcmd.PreRunner, analyticsClient analytics.Client) *appCommand {
	cmd := &cobra.Command{
		Use:         "app",
		Short:       "Manage ksqlDB apps.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &appCommand{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner, subcommandFlags),
		analyticsClient:               analyticsClient,
		prerunner:                     prerunner,
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
