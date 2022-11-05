package context

import (
	"strconv"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type listOut struct {
	Current    string `human:"Current" serialized:"current"`
	Name       string `human:"Name" serialized:"name"`
	Platform   string `human:"Platform" serialized:"platform"`
	Credential string `human:"Credential" serialized:"credential"`
}

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all contexts.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	list := output.NewList(cmd)
	for _, context := range c.Config.Contexts {
		isCurrent := context.Name == c.Config.CurrentContext

		current := ""
		if isCurrent {
			current = "*"
		}

		if output.GetFormat(cmd).IsSerialized() {
			current = strconv.FormatBool(isCurrent)
		}

		list.Add(&listOut{
			Current:    current,
			Name:       context.Name,
			Platform:   context.PlatformName,
			Credential: context.CredentialName,
		})
	}
	return list.Print()
}
