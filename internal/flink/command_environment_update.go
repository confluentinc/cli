package flink

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newEnvironmentUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a Flink environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.environmentUpdate,
	}

	cmd.Flags().String("url", "", `Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.`)
	cmd.Flags().String("client-key-path", "", `Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.`)
	cmd.Flags().String("client-cert-path", "", `Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.`)
	cmd.Flags().String("certificate-authority-path", "", `Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.`)
	cmd.Flags().String("defaults", "", "JSON string defining the environment's Flink application defaults, or path to a file to read defaults from (with .yml, .yaml or .json extension).")

	return cmd
}

func (c *command) environmentUpdate(cmd *cobra.Command, args []string) error {
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environmentName := args[0]

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
	var environment cmfsdk.PostEnvironment
	environment.Name = environmentName
	if defaultsParsed != nil {
		environment.FlinkApplicationDefaults = defaultsParsed
	}

	outputEnvironment, err := client.UpdateEnvironment(cmd.Context(), environment)
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
