package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type systemInfoOut struct {
	Version  string `human:"Version" serialized:"version"`
	Revision string `human:"Revision" serialized:"revision"`
}

type localSystemInformation struct {
	Status *localSystemInformationStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type localSystemInformationStatus struct {
	Version  *string `json:"version,omitempty" yaml:"version,omitempty"`
	Revision *string `json:"revision,omitempty" yaml:"revision,omitempty"`
}

func (c *command) newSystemInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "system-info",
		Short:       "Display CMF system information.",
		Args:        cobra.NoArgs,
		RunE:        c.systemInfo,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) systemInfo(cmd *cobra.Command, _ []string) error {
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	result, err := client.GetSystemInformation(c.createContext())
	if err != nil {
		return err
	}

	sysInfo := parseSystemInformation(result)

	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
		table.Add(&systemInfoOut{
			Version:  derefString(sysInfo.Status.Version),
			Revision: derefString(sysInfo.Status.Revision),
		})
		return table.Print()
	}

	return output.SerializedOutput(cmd, sysInfo)
}

func parseSystemInformation(raw map[string]interface{}) localSystemInformation {
	sysInfo := localSystemInformation{}

	statusMap, ok := raw["status"].(map[string]interface{})
	if !ok {
		return sysInfo
	}

	status := &localSystemInformationStatus{}
	if v, ok := statusMap["version"].(string); ok {
		status.Version = &v
	}
	if v, ok := statusMap["revision"].(string); ok {
		status.Revision = &v
	}
	sysInfo.Status = status

	return sysInfo
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
