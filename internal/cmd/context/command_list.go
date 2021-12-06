package context

import (
	"strconv"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	listFields       = []string{"Current", "Name", "Platform", "Credential"}
	structuredLabels = []string{"current", "name", "platform", "credential"}
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all contexts.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}

	output.AddFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	w, err := output.NewListOutputWriter(cmd, listFields, listFields, structuredLabels)
	if err != nil {
		return err
	}

	for _, ctx := range c.Config.Contexts {
		isHuman := w.GetOutputFormat() == output.Human
		row := newRow(isHuman, ctx, c.Config.CurrentContext)
		w.AddElement(row)
	}
	w.StableSort()

	return w.Out()
}

type row struct {
	Current    string
	Name       string
	Platform   string
	Credential string
}

func newRow(isHuman bool, ctx *v1.Context, current string) *row {
	isCurrent := ctx.Name == current

	return &row{
		Current:    formatCurrent(isHuman, isCurrent),
		Name:       ctx.Name,
		Platform:   ctx.PlatformName,
		Credential: ctx.CredentialName,
	}
}

func formatCurrent(isHuman, isCurrent bool) string {
	if isHuman {
		if isCurrent {
			return "*"
		} else {
			return ""
		}
	} else {
		return strconv.FormatBool(isCurrent)
	}
}
