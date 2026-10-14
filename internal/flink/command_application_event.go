package flink

import (
	"encoding/json"

	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type flinkApplicationEventOut struct {
	Name      string `human:"Name" serialized:"name"`
	Type      string `human:"Type" serialized:"type"`
	Timestamp string `human:"Timestamp" serialized:"timestamp"`
	Instance  string `human:"Instance" serialized:"instance"`
	Message   string `human:"Message" serialized:"message"`
}

func (c *command) newApplicationEventCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "event",
		Short:       "Manage Flink application events.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.AddCommand(c.newApplicationEventListCommand())

	return cmd
}

func convertSdkEventToLocalEvent(sdkEvent cmfsdk.FlinkApplicationEvent) LocalFlinkApplicationEvent {
	return LocalFlinkApplicationEvent{
		ApiVersion: sdkEvent.ApiVersion,
		Kind:       sdkEvent.Kind,
		Metadata: LocalEventMetadata{
			Name:                     sdkEvent.Metadata.Name,
			Uid:                      sdkEvent.Metadata.Uid,
			CreationTimestamp:        sdkEvent.Metadata.CreationTimestamp,
			FlinkApplicationInstance: sdkEvent.Metadata.FlinkApplicationInstance,
			Labels:                   sdkEvent.Metadata.Labels,
			Annotations:              sdkEvent.Metadata.Annotations,
		},
		Status: LocalEventStatus{
			Message: sdkEvent.Status.Message,
			Type:    sdkEvent.Status.Type,
			Data:    convertSdkEventData(sdkEvent.Status.Data),
		},
	}
}

// convertSdkEventData flattens the SDK's oneOf EventData (EventDataJobException |
// EventDataNewStatus) into a generic map so either variant serializes cleanly to
// JSON and YAML.
func convertSdkEventData(data *cmfsdk.EventData) *map[string]interface{} {
	if data == nil {
		return nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	var generic map[string]interface{}
	if err := json.Unmarshal(bytes, &generic); err != nil || len(generic) == 0 {
		return nil
	}
	return &generic
}
