package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *exporterCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all schema exporters.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *exporterCommand) list(cmd *cobra.Command, _ []string) error {
	type listDisplay struct {
		Exporter string
	}

	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	exporters, _, err := srClient.DefaultApi.GetExporters(ctx)
	if err != nil {
		return err
	}

	if len(exporters) > 0 {
		outputWriter, err := output.NewListOutputWriter(cmd, []string{"Exporter"}, []string{"Exporter"}, []string{"Exporter"})
		if err != nil {
			return err
		}
		for _, exporter := range exporters {
			outputWriter.AddElement(&listDisplay{
				Exporter: exporter,
			})
		}
		return outputWriter.Out()
	} else {
		utils.Println(cmd, errors.NoExporterMsg)
	}

	return nil
}
