package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newApplicationEventListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink application events.",
		Args:  cobra.NoArgs,
		RunE:  c.applicationEventList,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().String("application", "", "Name of the Flink application.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cobra.CheckErr(cmd.MarkFlagRequired("application"))

	return cmd
}

func (c *command) applicationEventList(cmd *cobra.Command, _ []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	application, err := cmd.Flags().GetString("application")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	events, err := client.ListApplicationEvents(c.createContext(), environment, application)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, event := range events {
			list.Add(&flinkApplicationEventOut{
				Name:      event.Metadata.GetName(),
				Type:      event.Status.GetType(),
				Timestamp: event.Metadata.GetCreationTimestamp(),
				Instance:  event.Metadata.GetFlinkApplicationInstance(),
				Message:   event.Status.GetMessage(),
			})
		}
		return list.Print()
	}

	localEvents := make([]LocalFlinkApplicationEvent, 0, len(events))
	for _, sdkEvent := range events {
		localEvents = append(localEvents, convertSdkEventToLocalEvent(sdkEvent))
	}

	return output.SerializedOutput(cmd, localEvents)
}
