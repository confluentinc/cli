package pipeline

import (
	"context"
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	describeFields            = []string{"Id", "Name", "State"}
	describeHumanRenames      = map[string]string{"Id": "ID"}
	describeStructuredRenames = map[string]string{"Id": "id", "Name": "name", "State": "state"}
)

func (c *command) newCreateCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new pipeline.",
		Args:  cobra.NoArgs,
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a new pipeline in Stream Designer",
				Code: `confluent pipeline create --name "test pipeline" --ksqldb-cluster "lkc-0000" --description "this is a test pipeline"`,
			},
		),
	}

	pcmd.AddKsqldbClusterFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("name", "", "Name of the pipeline.")
	cmd.Flags().String("description", "", "Description of the pipeline.")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("ksqldb-cluster")

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	// get flag values
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	ksqldbCluster, _ := cmd.Flags().GetString("ksqldb-cluster")

	// get kafka cluster
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	// validate ksql id
	ksqlReq := &schedv1.KSQLCluster{
		AccountId: c.EnvironmentId(),
		Id:        ksqldbCluster,
	}

	_, err = c.Client.KSQL.Describe(context.Background(), ksqlReq)
	if err != nil {
		return err
	}

	// validate sr id
	// todo: with devel, this srCluster returned is not the same as what's running in system test cluster
	//       hence creation is failing for system test account. I haven't tried other account yet, will debug
	//       later
	srCluster, err := c.Config.Context().SchemaRegistryCluster(cmd)
	if err != nil {
		return err
	}

	// call api
	pipeline, err := c.V2Client.CreatePipeline(c.EnvironmentId(), kafkaCluster.ID, name, description, ksqldbCluster, srCluster.Id)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, pipelineListFields, pipelineListHumanLabels, pipelineListStructuredLabels)
	if err != nil {
		return err
	}

	element := &Pipeline{Id: *pipeline.Id, Name: *pipeline.Spec.DisplayName, State: *pipeline.Status.State}
	outputWriter.AddElement(element)

	return outputWriter.Out()
}
