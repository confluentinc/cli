package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"

	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newCreateCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new pipeline.",
		Args:  cobra.NoArgs,
		RunE:  c.create,
	}

	cmd.Flags().String("name", "", "Name for new pipeline.")
	cmd.Flags().String("ksqldb-cluster", "", "KSQL DB cluster for new pipeline.")
	cmd.Flags().String("description", "", "Description for new pipeline.")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("ksqldb-cluster")

	return cmd
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	ksql, _ := cmd.Flags().GetString("ksqldb-cluster")

	kafka_cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	ksqlReq := &schedv1.KSQLCluster{
		AccountId: c.EnvironmentId(), 
		Id: ksql,
	}

	ksqlCluster, err := c.Client.KSQL.Describe(context.Background(), ksqlReq)
	if err != nil {
		return err
	}

	sr_cluster, err := c.Config.Context().SchemaRegistryCluster(cmd)
	if err != nil {
		return err
	}

	if kafka_cluster.ID != ksqlCluster.KafkaClusterId {
		utils.Println(cmd, "KSQL DB Cluster not in Kafka Cluster")
		return nil
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

	postBody, _ := json.Marshal(map[string]string{
		"name":                   name,
		"description":            description,
		"ksqlId":                 ksql,
		"connectEndpoint":        fmt.Sprintf("https://devel.cpdev.cloud/api/connect/v1/environments/%s/clusters/%s", c.Context.GetCurrentEnvironmentId(), kafka_cluster.ID),
		"kafkaClusterEndpoint":   kafka_cluster.Bootstrap,
		"ksqlEndpoint":           ksqlCluster.Endpoint,
		"schemaRegistryEndpoint": sr_cluster.SchemaRegistryEndpoint,
		"schemaRegistryId":       sr_cluster.Id,
	})
	bytesPostBody := bytes.NewBuffer(postBody)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://devel.cpdev.cloud/api/sd/v1/environments/%s/clusters/%s/pipelines", c.Context.GetCurrentEnvironmentId(), kafka_cluster.ID), bytesPostBody)
	if err != nil {
		return err
	}

	req.AddCookie(cookie)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(string(body)), &data)
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 && err == nil {
		utils.Println(cmd, "Created pipeline: "+data["id"].(string))
	} else {
		if err != nil {
			return err
		} else if body != nil {
			if data["title"] != "{}" {
				utils.Println(cmd, data["title"].(string))
			}
			utils.Println(cmd, data["action"].(string))
		}
	}

	return nil
}
