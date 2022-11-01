package pipeline

import (
	"context"

	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newCreateCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new pipeline.",
		Args:  cobra.NoArgs,
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a new Stream Designer pipeline",
				Code: `confluent pipeline create --name test-pipeline --ksql-cluster lksqlc-12345 --description "this is a test pipeline"`,
			},
		),
	}

	pcmd.AddKsqlClusterFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("name", "", "Name of the pipeline.")
	cmd.Flags().String("description", "", "Description of the pipeline.")
	pcmd.AddOutputFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	_ = cmd.MarkFlagRequired("ksql-cluster")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func (c *command) create(cmd *cobra.Command, _ []string) error {
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	ksqlCluster, _ := cmd.Flags().GetString("ksql-cluster")

	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	// validate ksql id
	ksqlReq := &schedv1.KSQLCluster{
		AccountId: c.EnvironmentId(),
		Id:        ksqlCluster,
	}

	if _, err = c.Client.KSQL.Describe(context.Background(), ksqlReq); err != nil {
		return err
	}

	// validate sr id
	srCluster, err := c.Config.Context().SchemaRegistryCluster(cmd)
	if err != nil {
		return err
	}

	pipeline, err := c.V2Client.CreatePipeline(c.EnvironmentId(), kafkaCluster.ID, name, description, ksqlCluster, srCluster.Id)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		Id:          pipeline.GetId(),
		Name:        pipeline.Spec.GetDisplayName(),
		Description: pipeline.Spec.GetDescription(),
		KsqlCluster: pipeline.Spec.KsqlCluster.GetId(),
		State:       pipeline.Status.GetState(),
		CreatedAt:   pipeline.Metadata.GetCreatedAt(),
		UpdatedAt:   pipeline.Metadata.GetUpdatedAt(),
	})
	return table.Print()
}
