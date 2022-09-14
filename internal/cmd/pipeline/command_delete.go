package pipeline

import (
	"encoding/json"
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
)

func (c *command) newDeleteCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "delete <pipeline-id>",
		Short:       "Delete a pipeline.",
		Args:        cobra.ExactArgs(1),
		RunE:        c.delete,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}
	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		utils.Println(cmd, "Could not get Kafka Cluster with error: "+err.Error())
		return nil
	}

	var client http.Client
	jar, err := cookiejar.New(nil)
	if err != nil {
		utils.Println(cmd, "Could not delete pipeline with error: "+err.Error())
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

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("https://devel.cpdev.cloud/api/sd/v1/environments/%s/clusters/%s/pipelines/%s", c.Context.GetCurrentEnvironmentId(), cluster.ID, args[0]), nil)
	if err != nil {
		utils.Println(cmd, "Could not delete pipeline: "+args[0]+" with error: "+err.Error())
		return nil
	}

	req.AddCookie(cookie)

	resp, err := client.Do(req)
	if err != nil {
		utils.Println(cmd, "Could not delete pipeline: "+args[0]+" with error: "+err.Error())
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.Println(cmd, "Could not delete pipeline: "+args[0]+" with error: "+err.Error())
		return nil
	}

	if resp.StatusCode == 202 && err == nil {
		utils.Println(cmd, "Deleted pipeline: "+args[0])
	} else {
		utils.Print(cmd, "Could not delete pipeline: "+args[0])
		var data map[string]interface{}
		err = json.Unmarshal([]byte(string(body)), &data)
		if err != nil {
			utils.Println(cmd, "Could not delete pipeline: "+args[0]+" with error: "+err.Error())
			return nil
		} else if body != nil {
			var data map[string]interface{}
			err = json.Unmarshal([]byte(string(body)), &data)
			if err != nil {
				utils.Println(cmd, "Could not delete pipeline: "+args[0]+" with error: "+err.Error())
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
