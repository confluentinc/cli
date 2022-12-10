package pipeline

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"regexp"

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
	cmd.Flags().String("source-code-file", "", "Path to a sql file containing pipeline source code.")
	cmd.Flags().StringSlice("secret", nil, "A secret name value mapping.")
	pcmd.AddOutputFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	_ = cmd.MarkFlagRequired("ksql-cluster")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	ksqlCluster, _ := cmd.Flags().GetString("ksql-cluster")
	sourceCodeFile, _ := cmd.Flags().GetString("source-code-file")
	secrets, _ := cmd.Flags().GetStringSlice("secret")

	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	// validate ksql id
	ksqlReq := &schedv1.KSQLCluster{
		AccountId: c.EnvironmentId(),
		Id:        ksqlCluster,
	}

	_, err = c.PrivateClient.KSQL.Describe(context.Background(), ksqlReq)
	if err != nil {
		return err
	}

	// validate sr id
	srCluster, err := c.Config.Context().SchemaRegistryCluster(cmd)
	if err != nil {
		return err
	}

	// read pipeline source code file if provided
	sourceCode := ""
	if sourceCodeFile != "" {
		fileContent, err := ioutil.ReadFile(sourceCodeFile)
		if err != nil {
			return err
		}
		sourceCode = string(fileContent)
	}

	// parse and construct secret mappings
	secretMappings, err := createSecretMappings(secrets)
	if err != nil {
		return err
	}

	pipeline, err := c.V2Client.CreatePipeline(c.EnvironmentId(), kafkaCluster.ID, name, description, sourceCode, &secretMappings, ksqlCluster, srCluster.Id)
	if err != nil {
		return err
	}

	element := &Pipeline{
		Id:          *pipeline.Id,
		Name:        *pipeline.Spec.DisplayName,
		Description: *pipeline.Spec.Description,
		KsqlCluster: pipeline.Spec.KsqlCluster.Id,
		State:       *pipeline.Status.State,
		CreatedAt:   *pipeline.Metadata.CreatedAt,
		UpdatedAt:   *pipeline.Metadata.UpdatedAt,
	}

	return output.DescribeObject(cmd, element, pipelineDescribeFields, pipelineDescribeHumanLabels, pipelineDescribeStructuredLabels)
}

func createSecretMappings(secrets []string) (map[string]string, error) {
	secretMappings := make(map[string]string)

	// The name of a secret may consist of 1-64 lowercase letters, uppercase letters, digits,
	// and the '_' (underscore) and may not begin with a digit.
	pattern := regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)=(.*)$`)

	for _, secret := range secrets {
		if pattern.MatchString(secret) {
			matches := pattern.FindStringSubmatch(secret)
			if len(matches[1]) > 6 {
				return nil, fmt.Errorf("secret name cannot exceeds 64 characters")
			}

			secretMappings[matches[1]] = matches[2]
		} else {
			return nil, fmt.Errorf("each secret must conform to the pattern of <name>=<value>")
		}
	}
	return secretMappings, nil
}
