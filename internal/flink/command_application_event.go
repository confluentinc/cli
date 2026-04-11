package flink

import (
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

type LocalFlinkApplicationEvent struct {
	ApiVersion string                   `json:"apiVersion" yaml:"apiVersion"`
	Kind       string                   `json:"kind" yaml:"kind"`
	Metadata   LocalEventMetadata       `json:"metadata" yaml:"metadata"`
	Status     LocalEventStatus         `json:"status" yaml:"status"`
}

type LocalEventMetadata struct {
	Name                     *string            `json:"name,omitempty" yaml:"name,omitempty"`
	Uid                      *string            `json:"uid,omitempty" yaml:"uid,omitempty"`
	CreationTimestamp        *string            `json:"creationTimestamp,omitempty" yaml:"creationTimestamp,omitempty"`
	FlinkApplicationInstance *string            `json:"flinkApplicationInstance,omitempty" yaml:"flinkApplicationInstance,omitempty"`
	Labels                   *map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations              *map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type LocalEventStatus struct {
	Message *string `json:"message,omitempty" yaml:"message,omitempty"`
	Type    *string `json:"type,omitempty" yaml:"type,omitempty"`
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
		},
	}
}
