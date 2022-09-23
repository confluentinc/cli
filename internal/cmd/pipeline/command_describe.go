package pipeline

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDescribeCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <pipeline-id>",
		Short: "Describe a pipeline with the option to save pipeline model.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
	}

	cmd.Flags().Bool("save-sql", false, "Save the pipeline model in a local file.")
	cmd.Flags().String("output-directory", "", "Path to save pipeline model.")

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	saveSql, _ := cmd.Flags().GetBool("save-sql")
	outputDir, _ := cmd.Flags().GetString("output-directory")

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

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://devel.cpdev.cloud/api/sd/v1/environments/%s/clusters/%s/pipelines/%s", c.Context.GetCurrentEnvironmentId(), cluster.ID, args[0]), nil)
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
		var pipeline Pipeline
		clusterLabels := []string{"Id", "Name", "State"}

		err = json.Unmarshal([]byte(string(body)), &pipeline)
		if err != nil {
			return err
		}

		outputWriter, err := output.NewListOutputWriter(cmd, clusterLabels, clusterLabels, clusterLabels)
		if err != nil {
			return err
		}

		outputWriter.AddElement(&pipeline)
		return outputWriter.Out()
	} else {
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
			return nil
		}
	}

	if saveSql {

		// TODO: Create GET /{pipeline_id}/describe API to export SQL file
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://devel.cpdev.cloud/api/sd/v1/environments/%s/clusters/%s/pipelines/%s/content", c.Context.GetCurrentEnvironmentId(), cluster.ID, args[0]), nil)
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
			filepath := filepath.Join(outputDir, args[0] + ".sql")
			out, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				return err
			}

			defer out.Close()
			_, err = out.Write(body)
			if err != nil {
				return err
			}
			utils.Println(cmd, "Saved sql file for pipeline: "+args[0])
		} else {
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
