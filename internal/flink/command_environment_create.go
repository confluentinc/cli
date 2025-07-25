package flink

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newEnvironmentCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Flink environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.environmentCreate,
	}

	cmd.Flags().String("kubernetes-namespace", "", "Kubernetes namespace to deploy Flink applications to.")
	cmd.Flags().String("defaults", "", "JSON string defining the environment's Flink application defaults, or path to a file to read defaults from (with .yml, .yaml or .json extension).")
	cmd.Flags().String("statement-defaults", "", "JSON string defining the environment's Flink statement defaults, or path to a file to read defaults from (with .yml, .yaml or .json extension).")
	cmd.Flags().String("compute-pool-defaults", "", "JSON string defining the environment's Flink compute pool defaults, or path to a file to read defaults from (with .yml, .yaml or .json extension).")

	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("kubernetes-namespace"))

	return cmd
}

func (c *command) environmentCreate(cmd *cobra.Command, args []string) error {
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environmentName := args[0]

	kubernetesNamespace, err := cmd.Flags().GetString("kubernetes-namespace")
	if err != nil {
		return err
	}

	// Read file contents or parse defaults if applicable
	var defaultsApplicationParsed, defaultsComputePoolParsed map[string]interface{}
	var defaultsStatementParsed cmfsdk.AllStatementDefaults1

	defaultsApplication, err := cmd.Flags().GetString("defaults")
	if err != nil {
		return fmt.Errorf("failed to read defaults application: %v", err)
	}

	defaultsComputePool, err := cmd.Flags().GetString("compute-pool-defaults")
	if err != nil {
		return fmt.Errorf("failed to read defaults compute pool: %v", err)
	}

	defaultsStatement, err := cmd.Flags().GetString("statement-defaults")
	if err != nil {
		return fmt.Errorf("failed to read defaults statement: %v", err)
	}

	if defaultsApplication != "" {
		if defaultsApplicationParsed, err = parseDefaultsAsGenericType[map[string]interface{}](defaultsApplication, "application"); err != nil {
			return err
		}
	}
	if defaultsComputePool != "" {
		if defaultsComputePoolParsed, err = parseDefaultsAsGenericType[map[string]interface{}](defaultsComputePool, "compute-pool"); err != nil {
			return err
		}
	}
	if defaultsStatement != "" {
		if defaultsStatementParsed, err = parseDefaultsAsGenericType[cmfsdk.AllStatementDefaults1](defaultsStatement, "statement"); err != nil {
			return err
		}
	}

	var postEnvironment cmfsdk.PostEnvironment
	postEnvironment.Name = environmentName
	postEnvironment.FlinkApplicationDefaults = &defaultsApplicationParsed
	postEnvironment.KubernetesNamespace = &kubernetesNamespace
	postEnvironment.StatementDefaults = &defaultsStatementParsed
	postEnvironment.ComputePoolDefaults = &defaultsComputePoolParsed

	sdkOutputEnvironment, err := client.CreateEnvironment(c.createContext(), postEnvironment)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		return printEnvironmentOutTable(cmd, sdkOutputEnvironment)
	}

	// Start with the top-level fields
localEnv := LocalEnvironment{
	Secrets:                  sdkOutputEnvironment.Secrets,
	Name:                     sdkOutputEnvironment.Name,
	CreatedTime:              sdkOutputEnvironment.CreatedTime,
	UpdatedTime:              sdkOutputEnvironment.UpdatedTime,
	FlinkApplicationDefaults: sdkOutputEnvironment.FlinkApplicationDefaults,
	KubernetesNamespace:      sdkOutputEnvironment.KubernetesNamespace,
	ComputePoolDefaults:      sdkOutputEnvironment.ComputePoolDefaults,
}

// Perform a deep copy for the nested StatementDefaults struct, handling nil pointers.
if sdkOutputEnvironment.StatementDefaults != nil {
	localDefaults1 := &LocalAllStatementDefaults1{}

	if sdkOutputEnvironment.StatementDefaults.Detached != nil {
		localDefaults1.Detached = &LocalStatementDefaults{
			FlinkConfiguration: sdkOutputEnvironment.StatementDefaults.Detached.FlinkConfiguration,
		}
	}

	if sdkOutputEnvironment.StatementDefaults.Interactive != nil {
		localDefaults1.Interactive = &LocalStatementDefaults{
			FlinkConfiguration: sdkOutputEnvironment.StatementDefaults.Interactive.FlinkConfiguration,
		}
	}

	localEnv.StatementDefaults = localDefaults1
}
	return output.SerializedOutput(cmd, localEnv)
}

func parseDefaultsAsGenericType[T any](input, label string) (T, error) {
	var out T
	var data []byte
	var err error

	ext := strings.ToLower(filepath.Ext(input))
	switch ext {
	case ".json":
		data, err = os.ReadFile(input)
		if err != nil {
			return out, fmt.Errorf("failed to read %s defaults JSON file: %w", label, err)
		}
		err = json.Unmarshal(data, &out)

	case ".yaml", ".yml":
		data, err = os.ReadFile(input)
		if err != nil {
			return out, fmt.Errorf("failed to read %s defaults YAML file: %w", label, err)
		}
		err = yaml.Unmarshal(data, &out)

	default:
		// inline JSON string
		err = json.Unmarshal([]byte(input), &out)
	}

	if err != nil {
		return out, fmt.Errorf("failed to parse %s defaults: %w", label, err)
	}
	return out, nil
}

func jsonMarshalHelper(v interface{}, label string) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal %s: %v", label, err)
	}
	return string(data), nil
}

func printEnvironmentOutTable(cmd *cobra.Command, outputEnvironment cmfsdk.Environment) error {
	table := output.NewTable(cmd)
	var defaultsApplicationOutput, defaultComputePoolOutput string
	var defaultsDetachedStatementOutput, defaultsInteractiveStatementOutput string
	var err error

	if defaultsApplicationOutput, err = jsonMarshalHelper(outputEnvironment.FlinkApplicationDefaults, "Flink application defaults"); err != nil {
		return err
	}
	if defaultComputePoolOutput, err = jsonMarshalHelper(outputEnvironment.ComputePoolDefaults, "Flink compute-pool defaults"); err != nil {
		return err
	}
	if defaultsDetachedStatementOutput, err = jsonMarshalHelper(outputEnvironment.GetStatementDefaults().Detached, "Flink detached statement defaults"); err != nil {
		return err
	}
	if defaultsInteractiveStatementOutput, err = jsonMarshalHelper(outputEnvironment.GetStatementDefaults().Interactive, "Flink interactive statement defaults"); err != nil {
		return err
	}

	table.Add(&flinkEnvironmentOutput{
		Name:                         outputEnvironment.Name,
		KubernetesNamespace:          outputEnvironment.KubernetesNamespace,
		FlinkApplicationDefaults:     defaultsApplicationOutput,
		ComputePoolDefaults:          defaultComputePoolOutput,
		DetachedStatementDefaults:    defaultsDetachedStatementOutput,
		InteractiveStatementDefaults: defaultsInteractiveStatementOutput,
		CreatedTime:                  outputEnvironment.CreatedTime.String(),
		UpdatedTime:                  outputEnvironment.UpdatedTime.String(),
	})
	return table.Print()
}
