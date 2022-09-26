package pipeline

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newActivateCommand(prerunner pcmd.PreRunner) *cobra.Command {
	return &cobra.Command{
		Use:   "activate <pipeline-id>",
		Short: "Request to activate a pipeline.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.activate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Request to activate a pipeline in Stream Designer",
				Code: `confluent pipeline activate pipe-12345`,
			},
		),
	}
}

func (c *command) activate(cmd *cobra.Command, args []string) error {
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

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://devel.cpdev.cloud/api/sd/v1/environments/%s/clusters/%s/pipelines/%s/activate", c.Context.GetCurrentEnvironmentId(), cluster.ID, args[0]), nil)
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
		utils.Println(cmd, "Activation request for pipeline: "+args[0]+" is accepted and in processing.")
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
