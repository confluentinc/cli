package flink

import (
	"slices"

	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newEnvironmentListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink environments.",
		Args:  cobra.NoArgs,
		RunE:  c.environmentList,
	}

	addPageSizeFlag(cmd)
	addCmfFlagSet(cmd)

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) environmentList(cmd *cobra.Command, _ []string) error {
	pageSize, err := getPageSize(cmd)
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkEnvironments, err := client.ListEnvironments(c.createContext(), pageSize)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		columns := []string{"Name", "CreatedTime", "UpdatedTime", "KubernetesNamespace"}
		// The column only appears while some environment is being deleted.
		if slices.ContainsFunc(sdkEnvironments, func(env cmfsdk.Environment) bool { return env.Metadata.GetDeletionRequestedTimestamp() != "" }) {
			columns = append(columns, "DeletionRequestedTime")
		}
		list.Filter(columns)
		for _, env := range sdkEnvironments {
			list.Add(&flinkEnvironmentOutput{
				Name:                  env.Name,
				KubernetesNamespace:   env.KubernetesNamespace,
				CreatedTime:           env.CreatedTime.String(),
				UpdatedTime:           env.UpdatedTime.String(),
				DeletionRequestedTime: env.Metadata.GetDeletionRequestedTimestamp(),
			})
		}
		return list.Print()
	}

	printableEnvs := make([]LocalEnvironment, 0, len(sdkEnvironments))
	for _, sdkEnv := range sdkEnvironments {
		localEnv := convertSdkEnvironmentToLocalEnvironment(sdkEnv)
		printableEnvs = append(printableEnvs, localEnv)
	}

	return output.SerializedOutput(cmd, printableEnvs)
}
