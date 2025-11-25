package flink

import (
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/spf13/cobra"
)

func (c *command) newDetachedSavepointDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink Detached Savepoint.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.detachedSavepointCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe a Flink savepoint named "my-savepoint".`,
				Code: `confluent flink detached-savepoint describe my-savepoint`,
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) detachedSavepointDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	detachedSavepoint, err := client.DescribeDetachedSavepoint(c.createContext(), name)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
		table.Add(&detachedSavepointOut{
			Name:              name,
			Path:              detachedSavepoint.Spec.GetPath(),
			Format:            detachedSavepoint.Spec.GetFormatType(),
			Limit:             detachedSavepoint.Spec.GetBackoffLimit(),
			CreationTimestamp: detachedSavepoint.Metadata.GetCreationTimestamp(),
			Uid:               detachedSavepoint.Metadata.GetUid(),
		})
		return table.Print()
	}
	localDetachedSavepoint := convertSdkSavepointToLocalSavepoint(detachedSavepoint)
	return output.SerializedOutput(cmd, localDetachedSavepoint)
}
