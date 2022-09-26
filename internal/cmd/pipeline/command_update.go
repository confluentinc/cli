package pipeline

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	sdv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newUpdateCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <pipeline-id>",
		Short: "Update an existing pipeline.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Request to update a pipeline in Stream Designer with a new name, description, and source code located in current local directory",
				Code: `confluent pipeline update pipe-12345 --name "NewPipeline" -- description "NewDescription" -- sql-file .`,
			},
		),
	}

	cmd.Flags().String("name", "", "New pipeline name.")
	cmd.Flags().String("description", "", "New pipeline description.")
	cmd.Flags().String("sql-file", "", "Path to the new pipeline model file.")

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	sqlFile, _ := cmd.Flags().GetString("sql-file")

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	if name == "" && description == "" && sqlFile == "" {
		utils.Println(cmd, "At least one field must be specified with --name, --description, or --sql-file")
		return nil
	}

	if name != "" || description != "" {
		updateBody := sdv1.NewSdV1PipelineUpdate()
		if name != "" {
			updateBody.SetName(name)
		}
		if description != "" {
			updateBody.SetDescription(description)
		}

		// call api
		_, resp, err := c.V2Client.UpdateSdPipeline(c.EnvironmentId(), cluster.ID, args[0], *updateBody)
		if err != nil {
			return err
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode == 200 && err == nil {
			utils.Println(cmd, "Updated pipeline: "+args[0])
		} else {
			utils.Print(cmd, "Could not update pipeline code: "+args[0])
			if err != nil {
				return err
			} else if body != nil {
				utils.Print(cmd, " with error: "+string(body))
				return nil
			}
		}
	}

	if sqlFile != "" {
		// get SQL content from filepath
		putBody, err := os.Open(sqlFile)
		if err != nil {
			return err
		}

		defer putBody.Close()

		// replace with PUT request body when minispec file is updated
		updateBody := sdv1.NewSdV1PipelineUpdate()

		// call PUT api when minispec file is updated
		_, resp, err := c.V2Client.UpdateSdPipeline(c.EnvironmentId(), cluster.ID, args[0], *updateBody)
		if err != nil {
			return err
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode == 200 && err == nil {
			utils.Println(cmd, "Replaced pipeline: "+args[0])
		} else {
			utils.Print(cmd, "Could not replace pipeline code: "+args[0])
			if err != nil {
				return err
			} else if body != nil {
				utils.Print(cmd, " with error: "+string(body))
				return nil
			}

		}
	}
	return nil
}
