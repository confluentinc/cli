package pipeline

import (
	"context"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/utils"
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
	}

	cmd.Flags().String("name", "", "Name for new pipeline.")
	cmd.Flags().String("ksqldb-cluster", "", "KSQL DB cluster for new pipeline.")
	cmd.Flags().String("description", "", "Description for new pipeline.")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("ksqldb-cluster")

	return cmd
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	// get flag values
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	ksql, _ := cmd.Flags().GetString("ksqldb-cluster")

	// get kafka cluster
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	// validate ksql id
	ksqlReq := &schedv1.KSQLCluster{
		AccountId: c.EnvironmentId(),
		Id:        ksql,
	}

	ksqlCluster, err := c.Client.KSQL.Describe(context.Background(), ksqlReq)
	if err != nil {
		return err
	}

	if kafkaCluster.ID != ksqlCluster.KafkaClusterId {
		utils.Println(cmd, "KSQL DB Cluster not in Kafka Cluster")
		return nil
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
	resp, _, err := c.V2Client.CreatePipeline(c.EnvironmentId(), kafkaCluster.ID, name, description, ksqlCluster.Id, srCluster.Id)
	if err != nil {
		return err
	}

	describePipeline := &Pipeline{Id: *resp.Id, Name: *resp.Spec.DisplayName, State: *resp.Status.State}

	return output.DescribeObject(cmd, describePipeline, describeFields, describeHumanRenames, describeStructuredRenames)
}
