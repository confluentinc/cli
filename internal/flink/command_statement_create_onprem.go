package flink

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	clierrors "github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/wait"
)

func (c *command) newStatementCreateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "create [name]",
		Short:       "Create a Flink SQL statement.",
		Long:        "Create a Flink SQL statement in Confluent Platform.",
		Args:        cobra.MaximumNArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.statementCreateOnPrem,
	}

	cmd.Flags().String("sql", "", "The Flink SQL statement.")
	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().String("compute-pool", "", "The compute pool name to execute the Flink SQL statement.")
	cmd.Flags().Uint16("parallelism", 1, "The parallelism the statement, default value is 1.")
	cmd.Flags().String("catalog", "", "The name of the default catalog.")
	cmd.Flags().String("database", "", "The name of the default database.")
	cmd.Flags().String("flink-configuration", "", "The file path to hold the Flink configuration for the statement.")
	pcmd.AddWaitFlag(cmd)
	pcmd.AddWaitTimeoutFlag(cmd, flinkStatementCreateWaitTimeout)
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("sql"))
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cobra.CheckErr(cmd.MarkFlagRequired("compute-pool"))

	return cmd
}

func (c *command) statementCreateOnPrem(cmd *cobra.Command, args []string) error {
	// Flink statement name can be automatically generated or provided by the user
	name := types.GenerateStatementNameForOnPrem()
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
	flinkConfiguration, err := c.readFlinkConfiguration(cmd)
	if err != nil {
		return err
	}
	statement := cmfsdk.Statement{
		ApiVersion: "cmf.confluent.io/v1",
		Kind:       "Statement",
		Metadata: cmfsdk.StatementMetadata{
			Name: name,
		},
		Spec: cmfsdk.StatementSpec{
			Statement:          sql,
			Properties:         &properties,
			FlinkConfiguration: &flinkConfiguration,
			ComputePoolName:    computePool,
			Parallelism:        cmfsdk.PtrInt32(int32(parallelism)),
			Stopped:            cmfsdk.PtrBool(false),
		},
	}
	shouldWait, err := cmd.Flags().GetBool("wait")
	if err != nil {
		return err
	}

	finalStatement, err := client.CreateStatement(c.createContext(), environment, statement)
	if err != nil {
		return err
	}

	if shouldWait {
		timeout, err := cmd.Flags().GetDuration("wait-timeout")
		if err != nil {
			return err
		}
		finalStatement, err = wait.PollPhases(cmd.Context(), wait.PhaseOptions[cmfsdk.Statement]{
			Fetch: func() (cmfsdk.Statement, error) {
				return client.GetStatement(c.createContext(), environment, name)
			},
			Phase:         func(s cmfsdk.Statement) string { return s.GetStatus().Phase },
			PendingPhases: flinkStatementPendingPhases,
			FailedPhases:  flinkStatementFailedPhases,
			Tick:          2 * time.Second,
			Timeout:       timeout,
		})
		if err != nil {
			switch {
			case errors.Is(err, wait.ErrFailed):
				status := finalStatement.GetStatus()
				return clierrors.NewErrorWithSuggestions(
					fmt.Sprintf(`statement "%s" entered failed phase %q: %s`, name, status.Phase, status.GetDetail()),
					fmt.Sprintf("Inspect the statement with `confluent flink statement describe %s`.", name),
				)
			case errors.Is(err, wait.ErrTimeout):
				return clierrors.NewErrorWithSuggestions(err.Error(), "Increase `--wait-timeout` or omit `--wait`.")
			default:
				// Fetch error or context cancellation — surface the underlying
				// cause unmodified; suggesting a longer timeout would mislead.
				return err
			}
		}
	}

	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
		table.Add(&statementOutOnPrem{
			CreationDate: finalStatement.Metadata.GetCreationTimestamp(),
			Name:         finalStatement.Metadata.GetName(),
			Statement:    finalStatement.Spec.GetStatement(),
			ComputePool:  finalStatement.Spec.GetComputePoolName(),
			Status:       finalStatement.Status.GetPhase(),
			StatusDetail: finalStatement.Status.GetDetail(),
			Parallelism:  finalStatement.Spec.GetParallelism(),
			Stopped:      finalStatement.Spec.GetStopped(),
			SqlKind:      finalStatement.Status.Traits.GetSqlKind(),
			AppendOnly:   finalStatement.Status.Traits.GetIsAppendOnly(),
			Bounded:      finalStatement.Status.Traits.GetIsBounded(),
		})
		return table.Print()
	}

	localStmt := convertSdkStatementToLocalStatement(finalStatement)
	return output.SerializedOutput(cmd, localStmt)
}

func (c *command) readFlinkConfiguration(cmd *cobra.Command) (map[string]string, error) {
	configFilePath, err := cmd.Flags().GetString("flink-configuration")
	if err != nil {
		return nil, err
	}

	flinkConfiguration := map[string]string{}
	if configFilePath != "" {
		var data []byte
		data, err = os.ReadFile(configFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read Flink configuration file: %v", err)
		}
		ext := filepath.Ext(configFilePath)
		switch ext {
		case ".json":
			err = json.Unmarshal(data, &flinkConfiguration)
		case ".yaml", ".yml":
			err = yaml.Unmarshal(data, &flinkConfiguration)
		default:
			return nil, clierrors.NewErrorWithSuggestions(fmt.Sprintf("unsupported file format: %s", ext), "Supported file formats are .json, .yaml, and .yml.")
		}
		if err != nil {
			return nil, err
		}
	}

	return flinkConfiguration, nil
}
