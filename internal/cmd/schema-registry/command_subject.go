package schema_registry

import (
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/go-printer"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"
)

type subjectCommand struct {
	*cobra.Command
	config   *config.Config
	ch       *pcmd.ConfigHelper
	srClient *srsdk.APIClient
}

// NewSubjectCommand returns the Cobra command for Schema Registry subject list
func NewSubjectCommand(config *config.Config, ch *pcmd.ConfigHelper, srClient *srsdk.APIClient) *cobra.Command {
	subjectCmd := &subjectCommand{
		Command: &cobra.Command{
			Use:   "subject",
			Short: "List subjects",
		},
		config:   config,
		ch:       ch,
		srClient: srClient,
	}
	subjectCmd.init()
	return subjectCmd.Command
}

func (c *subjectCommand) init() {

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subjects",
		Example: `
Retrieve all subjects available in a Schema Registry

::
		ccloud schema-registry subject list
`,
		RunE: c.list,
		Args: cobra.NoArgs,
	}
	c.AddCommand(cmd)
}

func (c *subjectCommand) list(cmd *cobra.Command, args []string) error {
	var listLabels = []string{"Subject"}
	var data [][]string
	type listDisplay struct {
		Subject string
	}
	srClient, ctx, err := GetApiClient(c.srClient, c.ch)
	if err != nil {

		return err
	}
	list, _, err := srClient.DefaultApi.List(ctx)
	if err != nil {
		return err
	}
	if len(list) > 0 {
		for _, l := range list {
			data = append(data, printer.ToRow(&listDisplay{
				Subject: l,
			}, listLabels))
		}
		printer.RenderCollectionTable(data, listLabels)
	} else {
		fmt.Println("No subjects")
	}
	return nil
}
