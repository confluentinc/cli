package flink

import (
	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newSavepointCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a Flink savepoint.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.savepointCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a Flink savepoint named "my-savepoint".`,
				Code: `confluent flink savepoint create statement "SELECT 1;" --path path-to-savepoint --environment env1`,
			},
		),
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().String("application", "", "The name of the Flink application to create the savepoint for.")
	cmd.Flags().String("statement", "", "The name of the Flink statement to create the savepoint for.")
	cmd.Flags().String("path", "", "The directory where the savepoint should be stored.")
	cmd.Flags().String("format", "CANONICAL", "The format of the savepoint. Defaults to CANONICAL.")
	cmd.Flags().Int("backoff-limit", 0, "Maximum number of retries before the snapshot is considered failed. Set to -1 for unlimited or 0 for no retries.")
	pcmd.AddOutputFlag(cmd)
	addCmfFlagSet(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cmd.MarkFlagsOneRequired("application", "statement")
	cmd.MarkFlagsMutuallyExclusive("application", "statement")

	return cmd
}

func (c *command) savepointCreate(cmd *cobra.Command, args []string) error {
	name := ""
	if len(args) == 1 {
		name = args[0]
	}

	application, err := cmd.Flags().GetString("application")
	if err != nil {
		return err
	}

	statement, err := cmd.Flags().GetString("statement")
	if err != nil {
		return err
	}

	path, err := cmd.Flags().GetString("path")
	if err != nil {
		return err
	}

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return err
	}

	limit, err := cmd.Flags().GetInt("backoff-limit")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	limit32 := int32(limit)
	savepoint := cmfsdk.Savepoint{
		ApiVersion: "cmf.confluent.io/v1",
		Kind:       "Savepoint",
		Spec: cmfsdk.SavepointSpec{
			BackoffLimit: &limit32,
			FormatType:   &format,
		},
		Status: &cmfsdk.SavepointStatus{
			Path: &path,
		},
	}
	if path != "" {
		savepoint.Spec.SetPath(path)
	}
	if name != "" {
		savepoint.Metadata.SetName(name)
	}
	var savepointCreated cmfsdk.Savepoint

	if application != "" {
		savepointCreated, err = client.CreateSavepointApplication(c.createContext(), savepoint, environment, application)
		if err != nil {
			return err
		}
	} else if statement != "" {
		savepointCreated, err = client.CreateSavepointStatement(c.createContext(), savepoint, environment, statement)
		if err != nil {
			return err
		}
	}

	table := output.NewTable(cmd)
	table.Add(&savepointOut{
		Name:         savepointCreated.Metadata.GetName(),
		Statement:    statement,
		Application:  application,
		Path:         savepointCreated.Spec.GetPath(),
		Format:       savepointCreated.Spec.GetFormatType(),
		BackoffLimit: savepointCreated.Spec.GetBackoffLimit(),
		Uid:          savepointCreated.Metadata.GetUid(),
		State:        savepointCreated.Status.GetState(),
	})
	return table.Print()
}
