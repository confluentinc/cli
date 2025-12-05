package flink

import (
	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newDetachedSavepointCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a Flink Detached Savepoint.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.detachedSavepointCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a Flink savepoint named "my-savepoint".`,
				Code: `confluent flink detached-savepoint create ds1 --path path1`,
			},
		),
	}

	cmd.Flags().String("path", "", " The path to the savepoint data.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)
	addCmfFlagSet(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("path"))

	return cmd
}

func (c *command) detachedSavepointCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	path, err := cmd.Flags().GetString("path")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	savepoint := cmfsdk.Savepoint{
		ApiVersion: "cmf.confluent.io/v1",
		Kind:       "Savepoint",
		Metadata: cmfsdk.SavepointMetadata{
			Name: &name,
		},
		Spec: cmfsdk.SavepointSpec{
			Path: &path,
		},
		Status: &cmfsdk.SavepointStatus{
			Path: &path,
		},
	}

	detachedSavepoint, _, err := client.DetachedSavepointsApi.CreateDetachedSavepoint(c.createContext()).Savepoint(savepoint).Execute()
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&detachedSavepointOut{
		Name:              detachedSavepoint.Metadata.GetName(),
		Path:              detachedSavepoint.Spec.GetPath(),
		Format:            detachedSavepoint.Spec.GetFormatType(),
		Limit:             detachedSavepoint.Spec.GetBackoffLimit(),
		CreationTimestamp: detachedSavepoint.Metadata.GetCreationTimestamp(),
		Uid:               detachedSavepoint.Metadata.GetUid(),
	})
	return table.Print()
}
