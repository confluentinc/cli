package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *customRuntimeCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List custom connector runtimes.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}
	pcmd.AddOutputFlag(cmd)
	return cmd
}

func (c *customRuntimeCommand) list(cmd *cobra.Command, _ []string) error {
	runtimes, err := c.V2Client.ListCustomConnectorRuntimes()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, runtime := range runtimes {
		out := &customRuntimeOut{
			Id:                             runtime.GetId(),
			CustomConnectPluginRuntimeName: runtime.GetCustomConnectPluginRuntimeName(),
			RuntimeAkVersion:               runtime.GetRuntimeAkVersion(),
			SupportedJavaVersions:          runtime.GetSupportedJavaVersions(),
			ProductMaturity:                runtime.GetProductMaturity(),
			Description:                    runtime.GetDescription(),
		}
		eol, ok := runtime.GetEndOfLifeAtOk()
		if ok {
			out.EndOfLifeAt = eol.String()
		}
		list.Add(out)
	}
	list.Sort(true)
	return list.Print()
}
