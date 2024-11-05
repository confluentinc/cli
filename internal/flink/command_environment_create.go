package flink

import (
	"encoding/json"
	"fmt"
	"os"
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
	var defaultsParsed map[string]interface{}
	defaults, err := cmd.Flags().GetString("defaults")
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	if defaults != "" {
		defaultsParsed = make(map[string]interface{})
		if strings.HasSuffix(defaults, ".json") {
			var data []byte
			data, err = os.ReadFile(defaults)
			if err != nil {
				return fmt.Errorf("failed to read defaults file: %v", err)
			}
			err = json.Unmarshal(data, &defaultsParsed)
		} else if strings.HasSuffix(defaults, ".yaml") || strings.HasSuffix(defaults, ".yml") {
			var data []byte
			data, err = os.ReadFile(defaults)
			if err != nil {
				return fmt.Errorf("failed to read defaults file: %v", err)
			}
			err = yaml.Unmarshal(data, &defaultsParsed)
		} else {
			err = json.Unmarshal([]byte(defaults), &defaultsParsed)
		}

		if err != nil {
			return fmt.Errorf("failed to parse defaults: %v", err)
		}
	}

	var postEnvironment cmfsdk.PostEnvironment
	postEnvironment.Name = environmentName
	if defaultsParsed != nil {
		postEnvironment.FlinkApplicationDefaults = defaultsParsed
	}
	postEnvironment.KubernetesNamespace = kubernetesNamespace

	outputEnvironment, err := client.CreateEnvironment(cmd.Context(), postEnvironment)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	var defaultsBytes []byte
	defaultsBytes, err = json.Marshal(outputEnvironment.FlinkApplicationDefaults)
	if err != nil {
		return fmt.Errorf("failed to marshal defaults: %s", err)
	}

	table.Add(&flinkEnvironmentOutput{
		Name:                     outputEnvironment.Name,
		KubernetesNamespace:      outputEnvironment.KubernetesNamespace,
		FlinkApplicationDefaults: string(defaultsBytes),
		CreatedTime:              outputEnvironment.CreatedTime.String(),
		UpdatedTime:              outputEnvironment.UpdatedTime.String(),
	})
	return table.Print()
}
