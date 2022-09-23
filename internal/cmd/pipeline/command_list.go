package pipeline

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type Pipeline struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

func (c *command) newListCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Display pipelines in the current environment and cluster.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, args []string) error {
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

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://devel.cpdev.cloud/api/sd/v1/environments/%s/clusters/%s/pipelines", c.Context.GetCurrentEnvironmentId(), cluster.ID), nil)
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
		var pipelines []Pipeline
		err = json.Unmarshal([]byte(string(body)), &pipelines)
		if err != nil {
			return err
		}

		clusterLabels := []string{"Id", "Name", "State"}

		outputWriter, err := output.NewListOutputWriter(cmd, clusterLabels, clusterLabels, clusterLabels)
		if err != nil {
			return err
		}

		for _, element := range pipelines {
			outputWriter.AddElement(&element)
		}
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
		}
	}

	return nil
}
