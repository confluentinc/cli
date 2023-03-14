package pipeline

import (
	"strings"
	"time"

	"github.com/spf13/cobra"

	streamdesignerv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	launchdarkly "github.com/confluentinc/cli/internal/pkg/featureflags"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type humanOut struct {
	Id                  string    `human:"ID"`
	Name                string    `human:"Name"`
	Description         string    `human:"Description"`
	KsqlCluster         string    `human:"KSQL Cluster"`
	SecretNames         string    `human:"Secret Names,omitempty"`
	ActivationPrivilege bool      `human:"Activation Privilege"`
	State               string    `human:"State"`
	CreatedAt           time.Time `human:"Created At"`
	UpdatedAt           time.Time `human:"Updated At"`
}

type serializedOut struct {
	Id                  string    `serialized:"id"`
	Name                string    `serialized:"name"`
	Description         string    `serialized:"description"`
	KsqlCluster         string    `serialized:"ksql_cluster"`
	SecretNames         []string  `serialized:"secret_names,omitempty"`
	ActivationPrivilege bool      `serialized:"activation_privilege"`
	State               string    `serialized:"state"`
	CreatedAt           time.Time `serialized:"created_at"`
	UpdatedAt           time.Time `serialized:"updated_at"`
}

var (
	secretMappingWithoutEmptyValue = `^([a-zA-Z_][a-zA-Z0-9_]*)=(.+)$`
	secretMappingWithEmptyValue    = `^([a-zA-Z_][a-zA-Z0-9_]*)=(.*)$`
)

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner pcmd.PreRunner
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "pipeline",
		Short:       "Manage Stream Designer pipelines.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner),
		prerunner:                     prerunner,
	}

	dc := dynamicconfig.New(cfg, nil, nil)
	_ = dc.ParseFlagsIntoConfig(cmd)
	enableSourceCode := launchdarkly.Manager.BoolVariation("cli.stream_designer.source_code.enable", dc.Context(), v1.CliLaunchDarklyClient, true, false)

	c.AddCommand(c.newActivateCommand())
	c.AddCommand(c.newCreateCommand(enableSourceCode))
	c.AddCommand(c.newDeactivateCommand())
	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newSaveCommand(enableSourceCode))
	c.AddCommand(c.newUpdateCommand(enableSourceCode))

	return c.Command
}

func printTable(cmd *cobra.Command, pipeline streamdesignerv1.SdV1Pipeline) error {
	table := output.NewTable(cmd)
	secrets := getOrderedSecretNames(pipeline.Spec.Secrets)

	if output.GetFormat(cmd) == output.Human {
		table.Add(&humanOut{
			Id:                  pipeline.GetId(),
			Name:                pipeline.Spec.GetDisplayName(),
			Description:         pipeline.Spec.GetDescription(),
			KsqlCluster:         pipeline.Spec.KsqlCluster.GetId(),
			SecretNames:         strings.Join(secrets, ", "),
			ActivationPrivilege: pipeline.Spec.GetActivationPrivilege(),
			State:               pipeline.Status.GetState(),
			CreatedAt:           pipeline.Metadata.GetCreatedAt(),
			UpdatedAt:           pipeline.Metadata.GetUpdatedAt(),
		})
	} else {
		table.Add(&serializedOut{
			Id:                  pipeline.GetId(),
			Name:                pipeline.Spec.GetDisplayName(),
			Description:         pipeline.Spec.GetDescription(),
			KsqlCluster:         pipeline.Spec.KsqlCluster.GetId(),
			SecretNames:         secrets,
			ActivationPrivilege: pipeline.Spec.GetActivationPrivilege(),
			State:               pipeline.Status.GetState(),
			CreatedAt:           pipeline.Metadata.GetCreatedAt(),
			UpdatedAt:           pipeline.Metadata.GetUpdatedAt(),
		})
	}

	return table.Print()
}
