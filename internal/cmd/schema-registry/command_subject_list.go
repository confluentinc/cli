package schemaregistry

import (
	"fmt"

	"github.com/antihax/optional"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func (c *subjectCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subjects.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Retrieve all subjects available in a Schema Registry:",
				Code: fmt.Sprintf("%s schema-registry subject list", version.CLIName),
			},
		),
	}

	cmd.Flags().BoolP("deleted", "D", false, "View the deleted subjects.")
	cmd.Flags().String("prefix", ":*:", "Subject prefix.")
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)

	return cmd
}

func (c *subjectCommand) list(cmd *cobra.Command, _ []string) error {
	type listDisplay struct {
		Subject string
	}

	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	deleted, err := cmd.Flags().GetBool("deleted")
	if err != nil {
		return err
	}

	subjectPrefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	listOpts := srsdk.ListOpts{Deleted: optional.NewBool(deleted), SubjectPrefix: optional.NewString(subjectPrefix)}
	list, _, err := srClient.DefaultApi.List(ctx, &listOpts)
	if err != nil {
		return err
	}
	if len(list) > 0 {
		outputWriter, err := output.NewListOutputWriter(cmd, []string{"Subject"}, []string{"Subject"}, []string{"subject"})
		if err != nil {
			return err
		}

		for _, l := range list {
			outputWriter.AddElement(&listDisplay{
				Subject: l,
			})
		}
		return outputWriter.Out()
	} else {
		utils.Println(cmd, errors.NoSubjectsMsg)
	}

	return nil
}
