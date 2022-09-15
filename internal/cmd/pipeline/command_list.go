package pipeline

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type Pipeline struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

func (c *command) newListCommand(prerunner pcmd.PreRunner) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Display pipelines in the current environment and cluster.",
		Args:  cobra.ExactArgs(0),
		RunE:  c.list,
	}
}

func (c *command) list(cmd *cobra.Command, args []string) error {
	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		utils.Println(cmd, "Could not get Kafka Cluster with error: "+err.Error())
		return err
	}

	var client http.Client
	jar, err := cookiejar.New(nil)
	if err != nil {
		utils.Println(cmd, "Could not list pipelines with error: "+err.Error())
		return nil
	}

	client = http.Client{
		Jar: jar,
	}

	cookie := &http.Cookie{
		Name:   "auth_token",
		Value:  c.State.AuthToken,
		MaxAge: 300,
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://devel.cpdev.cloud/api/sd/v1/environments/%s/clusters/%s/pipelines", c.Context.GetCurrentEnvironmentId(), cluster.ID), nil)
	if err != nil {
		utils.Println(cmd, "Could not list pipelines with error: "+err.Error())
		return nil
	}

	req.AddCookie(cookie)

	resp, err := client.Do(req)
	if err != nil {
		utils.Println(cmd, "Could not list pipelines with error: "+err.Error())
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Println(cmd, "Could not list pipelines with error: "+err.Error())
		return nil
	}

	if resp.StatusCode == 200 && err == nil {
		var pipelines []Pipeline
		err = json.Unmarshal([]byte(string(body)), &pipelines)
		if err != nil {
			utils.Println(cmd, "Could not list pipelines with error: "+err.Error())
			return nil
		}

		clusterLabels := []string{"id", "name", "state"}
		var out [][]string

		for _, element := range pipelines {
			out = append(out, []string{element.Id,
				element.Name,
				element.State})
		}

		tablePrinter := tablewriter.NewWriter(os.Stdout)
		tablePrinter.SetAutoWrapText(false)
		tablePrinter.SetAutoFormatHeaders(false)
		tablePrinter.SetHeader(clusterLabels)
		tablePrinter.AppendBulk(out)
		tablePrinter.SetBorder(false)
		tablePrinter.Render()
	} else {
		if err != nil {
			utils.Println(cmd, "Could not list pipelines with error: "+err.Error())
		} else if body != nil {
			var data map[string]interface{}
			err = json.Unmarshal([]byte(string(body)), &data)
			if err != nil {
				utils.Println(cmd, "Could not list pipelines with error: "+err.Error())
				return nil
			}
			if data["title"] != "{}" {
				utils.Println(cmd, data["title"].(string))
			}
			utils.Println(cmd, data["action"].(string))
		}
	}

	return nil
}
