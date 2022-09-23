package pipeline

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
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
		utils.Println(cmd, "Could not describe pipeline with error:"+err.Error())
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
		utils.Println(cmd, "Could not describe pipeline: "+args[0]+" with error: "+err.Error())
		return nil
	}

	req.AddCookie(cookie)

	resp, err := client.Do(req)
	if err != nil {
		utils.Println(cmd, "Could not describe pipeline: "+args[0]+" with error: "+err.Error())
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Println(cmd, "Could not describe pipeline: "+args[0]+" with error: "+err.Error())
		return nil
	}

	if resp.StatusCode == 200 && err == nil {
		clusterLabels := []string{"id", "name", "state"}
		var out [][]string

		var data map[string]interface{}
		err = json.Unmarshal([]byte(string(body)), &data)
		if err != nil {
			utils.Println(cmd, "Could not describe pipeline: "+args[0]+" with error: "+err.Error())
			return nil
		}
		out = append(out, []string{data["id"].(string),
			data["name"].(string),
			data["state"].(string)})

		tablePrinter := tablewriter.NewWriter(os.Stdout)
		tablePrinter.SetAutoWrapText(false)
		tablePrinter.SetAutoFormatHeaders(false)
		tablePrinter.SetHeader(clusterLabels)
		tablePrinter.AppendBulk(out)
		tablePrinter.SetBorder(false)
		tablePrinter.Render()
	} else {
		if err != nil {
			utils.Print(cmd, "Could not describe pipeline: "+args[0]+" with error: "+err.Error())
		} else if body != nil {
			var data map[string]interface{}
			err = json.Unmarshal([]byte(string(body)), &data)
			if err != nil {
				utils.Println(cmd, "Could not describe pipeline: "+args[0]+" with error: "+err.Error())
				return nil
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
			utils.Println(cmd, "Could not save sql file for pipeline: "+args[0]+" with error: "+err.Error())
			return nil
		}

		req.AddCookie(cookie)

		resp, err := client.Do(req)
		if err != nil {
			utils.Println(cmd, "Could not save sql file for pipeline: "+args[0]+" with error: "+err.Error())
			return nil
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			utils.Println(cmd, "Could not save sql file for pipeline: "+args[0]+" with error: "+err.Error())
			return nil
		}

		if resp.StatusCode == 200 && err == nil {
			filepath, err := fmt.Println(filepath.Join(outputDir, args[0] + ".sql"))

			out, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				utils.Println(cmd, "Could not save sql file for pipeline: "+args[0]+" with error: "+err.Error())
				return nil
			}

			defer out.Close()
			_, err = out.Write(body)
			if err != nil {
				utils.Println(cmd, "Could not save sql file for pipeline: "+args[0]+" with error: "+err.Error())
				return nil
			}
			utils.Println(cmd, "Saved sql file for pipeline: "+args[0])
		} else {
			if err != nil {
				utils.Print(cmd, "Could not save sql file for pipeline: "+args[0]+" with error: "+err.Error())
			} else if body != nil {
				var data map[string]interface{}
				err = json.Unmarshal([]byte(string(body)), &data)
				if err != nil {
					utils.Println(cmd, "Could not save sql file for pipeline: "+args[0]+" with error: "+err.Error())
					return nil
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
