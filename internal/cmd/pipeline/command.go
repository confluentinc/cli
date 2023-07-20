package pipeline

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	streamdesignerv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type humanOut struct {
	Id                    string    `human:"ID"`
	Name                  string    `human:"Name"`
	Description           string    `human:"Description"`
	KsqlCluster           string    `human:"KSQL Cluster,omitempty"`
	SchemaRegistryCluster string    `human:"Schema Registry Cluster,omitempty"`
	SecretNames           string    `human:"Secret Names,omitempty"`
	ActivationPrivilege   bool      `human:"Activation Privilege"`
	State                 string    `human:"State"`
	CreatedAt             time.Time `human:"Created At"`
	UpdatedAt             time.Time `human:"Updated At"`
}

type serializedOut struct {
	Id                    string    `serialized:"id"`
	Name                  string    `serialized:"name"`
	Description           string    `serialized:"description"`
	KsqlCluster           string    `serialized:"ksql_cluster"`
	SchemaRegistryCluster string    `serialized:"schema_registry_cluster"`
	SecretNames           []string  `serialized:"secret_names,omitempty"`
	ActivationPrivilege   bool      `serialized:"activation_privilege"`
	State                 string    `serialized:"state"`
	CreatedAt             time.Time `serialized:"created_at"`
	UpdatedAt             time.Time `serialized:"updated_at"`
}

var (
	secretMappingWithoutEmptyValue = `^([a-zA-Z_][a-zA-Z0-9_]*)=(.+)$`
	secretMappingWithEmptyValue    = `^([a-zA-Z_][a-zA-Z0-9_]*)=(.*)$`
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "pipeline",
		Short:       "Manage Stream Designer pipelines.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newActivateCommand())
	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeactivateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newSaveCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func printTable(cmd *cobra.Command, pipeline streamdesignerv1.SdV1Pipeline) error {
	table := output.NewTable(cmd)
	secrets := getOrderedSecretNames(pipeline.Spec.Secrets)

	if output.GetFormat(cmd) == output.Human {
		table.Add(&humanOut{
			Id:                    pipeline.GetId(),
			Name:                  pipeline.Spec.GetDisplayName(),
			Description:           pipeline.Spec.GetDescription(),
			KsqlCluster:           pipeline.Spec.KsqlCluster.GetId(),
			SchemaRegistryCluster: pipeline.Spec.StreamGovernanceCluster.GetId(),
			SecretNames:           strings.Join(secrets, ", "),
			ActivationPrivilege:   pipeline.Spec.GetActivationPrivilege(),
			State:                 pipeline.Status.GetState(),
			CreatedAt:             pipeline.Metadata.GetCreatedAt(),
			UpdatedAt:             pipeline.Metadata.GetUpdatedAt(),
		})
	} else {
		table.Add(&serializedOut{
			Id:                    pipeline.GetId(),
			Name:                  pipeline.Spec.GetDisplayName(),
			Description:           pipeline.Spec.GetDescription(),
			KsqlCluster:           pipeline.Spec.KsqlCluster.GetId(),
			SchemaRegistryCluster: pipeline.Spec.StreamGovernanceCluster.GetId(),
			SecretNames:           secrets,
			ActivationPrivilege:   pipeline.Spec.GetActivationPrivilege(),
			State:                 pipeline.Status.GetState(),
			CreatedAt:             pipeline.Metadata.GetCreatedAt(),
			UpdatedAt:             pipeline.Metadata.GetUpdatedAt(),
		})
	}

	return table.Print()
}

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validArgsMultiple(cmd, args)
}

func (c *command) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompletePipelines()
}

func (c *command) autocompletePipelines() []string {
	pipelines, err := c.getPipelines()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(pipelines))
	for i, pipeline := range pipelines {
		suggestions[i] = fmt.Sprintf("%s\t%s", pipeline.GetId(), pipeline.Spec.GetDisplayName())
	}
	return suggestions
}

func (c *command) getPipelines() ([]streamdesignerv1.SdV1Pipeline, error) {
	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return nil, err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListPipelines(environmentId, cluster.ID)
}
