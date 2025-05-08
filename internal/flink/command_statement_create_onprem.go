package flink

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/retry"
)

func (c *command) newStatementCreateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "create [name]",
		Short:       "Create a Flink SQL statement in Confluent Platform.",
		Args:        cobra.MaximumNArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.statementCreateOnPrem,
	}

	cmd.Flags().String("sql", "", "The Flink SQL statement.")
	cmd.Flags().String("environment", "", "Name of the environment to create the Flink SQL statement.")
	cmd.Flags().String("compute-pool", "", "The compute pool name to execute the Flink SQL statement.")
	cmd.Flags().Uint16("parallelism", 4, "The parallelism the statement, default value is 4.")
	cmd.Flags().String("catalog", "", "The name of the default catalog.")
	cmd.Flags().String("database", "", "The name of the default database.")
	cmd.Flags().String("flink-configuration", "", "The file path to hold the Flink configuration for the statement.")
	cmd.Flags().Bool("wait", false, "Boolean flag to block until the statement is running or has failed.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("sql"))
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cobra.CheckErr(cmd.MarkFlagRequired("compute-pool"))

	return cmd
}

func (c *command) statementCreateOnPrem(cmd *cobra.Command, args []string) error {
	// Flink statement name can be automatically generated or provided by the user
	name := types.GenerateStatementName()
	if len(args) == 1 {
		name = args[0]
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sql, err := cmd.Flags().GetString("sql")
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	computePool, err := cmd.Flags().GetString("compute-pool")
	if err != nil {
		return err
	}

	parallelism, err := cmd.Flags().GetUint16("parallelism")
	if err != nil {
		return err
	}

	catalog, err := cmd.Flags().GetString("catalog")
	if err != nil {
		return err
	}

	database, err := cmd.Flags().GetString("database")
	if err != nil {
		return err
	}

	properties := map[string]string{}
	if database != "" {
		properties["sql.current-database"] = database
	}
	if catalog != "" {
		properties["sql.current-catalog"] = catalog
	}

	configFilePath, err := cmd.Flags().GetString("flink-configuration")
	if err != nil {
		return err
	}

	var flinkConfiguration = map[string]string{}
	if configFilePath != "" {
		// Read configuration file contents
		data, err := os.ReadFile(configFilePath)
		if err != nil {
			return fmt.Errorf("failed to read Flink configuration file: %v", err)
		}
		ext := filepath.Ext(configFilePath)
		switch ext {
		case ".json":
			err = json.Unmarshal(data, &flinkConfiguration)
		case ".yaml", ".yml":
			err = yaml.Unmarshal(data, &flinkConfiguration)
		default:
			return errors.NewErrorWithSuggestions(fmt.Sprintf("unsupported file format: %s", ext), "Supported file formats are .json, .yaml, and .yml.")
		}
	}

	statement := cmfsdk.Statement{
		ApiVersion: "cmf.confluent.io/v1",
		Kind:       "Statement",
		Metadata: cmfsdk.StatementMetadata{
			Name: name,
		},
		Spec: cmfsdk.StatementSpec{
			Statement:          sql,
			Properties:         properties,
			FlinkConfiguration: flinkConfiguration,
			ComputePoolName:    computePool,
			Parallelism:        int32(parallelism),
			Stopped:            false,
		},
	}

	wait, err := cmd.Flags().GetBool("wait")
	if err != nil {
		return err
	}

	outputStatement, err := client.CreateStatement(c.createContext(), environment, statement)
	if err != nil {
		return err
	}

	// CreateStatement() is async API, add wait logic below when the statement is still PENDING
	if wait {
		err := retry.Retry(time.Second*2, time.Minute, func() error {
			statement, err = client.GetStatement(c.createContext(), environment, name)
			if err != nil {
				return err
			}

			if statement.Status.Phase == "PENDING" {
				return fmt.Errorf(`statement phase is "%s"`, statement.Status.Phase)
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	table := output.NewTable(cmd)
	table.Add(&statementOutOnPrem{
		CreationDate: time.Now(),
		Name:         outputStatement.Metadata.Name,
		Statement:    outputStatement.Spec.Statement,
		ComputePool:  outputStatement.Spec.ComputePoolName,
		Status:       outputStatement.Status.Phase,
		StatusDetail: outputStatement.Status.Detail,
		Properties:   outputStatement.Spec.Properties,
	})
	table.Filter([]string{"CreationTime", "Name", "Statement", "ComputePool", "Status", "StatusDetail", "Properties"})

	return table.Print()
}
