package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newDetachedSavepointListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink detached savepoint.",
		RunE:  c.detachedSavepointList,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List Flink detached savepoints with filter filter1.`,
				Code: `confluent flink detached-savepoint list --filter filter1`,
			},
		),
	}

	cmd.Flags().String("filter", "", "A filter expression to filter the list of detached savepoints.")

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)
	addCmfFlagSet(cmd)

	return cmd
}

func (c *command) detachedSavepointList(cmd *cobra.Command, args []string) error {
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		return err
	}

	detachedSavepoints, err := client.ListDetachedSavepoint(c.createContext(), filter)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, detachedSavepoint := range detachedSavepoints {
			list.Add(&detachedSavepointOut{
				Name:              detachedSavepoint.Metadata.GetName(),
				Path:              detachedSavepoint.Spec.GetPath(),
				Format:            detachedSavepoint.Spec.GetFormatType(),
				Limit:             detachedSavepoint.Spec.GetBackoffLimit(),
				CreationTimestamp: detachedSavepoint.Metadata.GetCreationTimestamp(),
				Uid:               detachedSavepoint.Metadata.GetUid(),
			})
		}
		return list.Print()
	}

	detachedSavepointsSdk := make([]LocalSavepoint, 0, len(detachedSavepoints))
	for _, sdksavepoint := range detachedSavepoints {
		savepoint := convertSdkDetachedSavepointToLocalSavepoint(sdksavepoint)
		detachedSavepointsSdk = append(detachedSavepointsSdk, savepoint)
	}

	return output.SerializedOutput(cmd, detachedSavepointsSdk)
}
