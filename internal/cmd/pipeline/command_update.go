package pipeline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newUpdateCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <pipeline-id>",
		Short: "Update an existing pipeline.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
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

	var client http.Client
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

	client = http.Client{
		Jar: jar,
	}

	cookie := &http.Cookie{
		Name:   "auth_token",
		Value:  c.State.AuthToken,
		MaxAge: 300,
	}

	if name == "" && description == "" && sqlFile == "" {
		utils.Println(cmd, "At least one field must be specified with --name, --description, or --sql-file")
		return nil
	}

	if name != "" || description != "" {
		postMap := make(map[string]string)
		if name != "" {
			postMap["name"] = name
		}
		if description != "" {
			postMap["description"] = description
		}

		postBody, _ := json.Marshal(postMap)
		bytesPostBody := bytes.NewBuffer(postBody)

		req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("https://devel.cpdev.cloud/api/sd/v1/environments/%s/clusters/%s/pipelines/%s", c.Context.GetCurrentEnvironmentId(), cluster.ID, args[0]), bytesPostBody)
		if err != nil {
			return err
		}

		req.AddCookie(cookie)
		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
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
		putBody, err := os.Open(sqlFile)
		if err != nil {
			return err
		}

		defer putBody.Close()

		// TODO: Modify PUT /{pipeline_id}/content API with a new @Consumes SQL file to import SQL file
		req, err := http.NewRequest("PUT", fmt.Sprintf("https://devel.cpdev.cloud/api/sd/v1/environments/%s/clusters/%s/pipelines/%s/content", c.Context.GetCurrentEnvironmentId(), cluster.ID, args[0]), putBody)
		if err != nil {
			return err
		}

		req.AddCookie(cookie)

		resp, err := client.Do(req)
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
				var data map[string]interface{}
				err = json.Unmarshal([]byte(string(body)), &data)
				if err != nil {
					return err
				}
				if data["title"] != "{}" {
					utils.Println(cmd, data["title"].(string))
				}
				utils.Println(cmd, data["action"].(string))
			}
		}
	}

	return nil
}
