package flink

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
)

type flinkApplicationSummaryOut struct {
	Name        string `human:"Name" serialized:"name"`
	Environment string `human:"Environment" serialized:"environment"`
	JobName     string `human:"Job Name" serialized:"job_name"`
	JobStatus   string `human:"Job Status" serialized:"job_status"`
}

func (c *command) newApplicationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "application",
		Short:       "Manage Flink applications.",
		Aliases:     []string{"app"},
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.AddCommand(c.newApplicationCreateCommand())
	cmd.AddCommand(c.newApplicationDeleteCommand())
	cmd.AddCommand(c.newApplicationDescribeCommand())
	cmd.AddCommand(c.newApplicationListCommand())
	cmd.AddCommand(c.newApplicationUpdateCommand())
	cmd.AddCommand(c.newApplicationWebUiForwardCommand())

	return cmd
}

func readApplicationResourceFile(resourceFilePath string) (cmfsdk.FlinkApplication, error) {
	data, err := os.ReadFile(resourceFilePath)
	if err != nil {
		return cmfsdk.FlinkApplication{}, fmt.Errorf("failed to read file: %v", err)
	}

	var genericData map[string]interface{}
	ext := filepath.Ext(resourceFilePath)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &genericData)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &genericData)
	default:
		return cmfsdk.FlinkApplication{}, errors.NewErrorWithSuggestions(fmt.Sprintf("unsupported file format: %s", ext), "Supported file formats are .json, .yaml, and .yml.")
	}
	if err != nil {
		return cmfsdk.FlinkApplication{}, fmt.Errorf("failed to parse input file: %w", err)
	}

	jsonBytes, err := json.Marshal(genericData)
	if err != nil {
		return cmfsdk.FlinkApplication{}, fmt.Errorf("failed to marshal intermediate data: %w", err)
	}

	var sdkApplication cmfsdk.FlinkApplication
	if err = json.Unmarshal(jsonBytes, &sdkApplication); err != nil {
		return cmfsdk.FlinkApplication{}, fmt.Errorf("failed to bind data to FlinkApplication model: %w", err)
	}

	return sdkApplication, nil
}

func convertSdkApplicationToLocalApplication(sdkApplication cmfsdk.FlinkApplication) LocalFlinkApplication {
	return LocalFlinkApplication{
		ApiVersion: sdkApplication.ApiVersion,
		Kind:       sdkApplication.Kind,
		Metadata:   sdkApplication.Metadata,
		Spec:       sdkApplication.Spec,
		Status:     sdkApplication.Status,
	}
}
