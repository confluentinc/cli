package kafka

import (
	"context"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/kafkarest"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func initKafkaRest(c *pcmd.AuthenticatedCLICommand, cmd *cobra.Command) (*kafkarestv3.APIClient, context.Context, string, error) {
	url, err := getKafkaRestUrl(cmd)
	if err != nil { // require the flag
		return nil, nil, "", err
	}

	if ccloudv2.IsCCloudURL(url, c.Config.IsTest) {
		output.ErrPrintf(c.Config.EnableColor, "[WARN] This is a Confluent Platform command. Confluent Cloud URLs are not supported.\n")
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return nil, nil, "", err
	}
	kafkaRestClient := kafkaREST.Client
	SetServerURL(cmd, kafkaRestClient, url)

	clusters, httpResp, err := kafkaRestClient.ClusterV3Api.ClustersGet(kafkaREST.Context)
	if err != nil {
		return nil, nil, "", kafkarest.NewError(kafkaRestClient.GetConfig().BasePath, err, httpResp)
	}
	if len(clusters.Data) == 0 {
		return nil, nil, "", errors.NewErrorWithSuggestions(errors.NoClustersFoundErrorMsg, errors.NoClustersFoundSuggestions)
	}

	return kafkaRestClient, kafkaREST.Context, clusters.Data[0].ClusterId, nil
}

// Used for on-prem KafkaRest commands
// Embedded KafkaRest uses /kafka/v3 and standalone uses /v3
// Relying on users to include the /kafka in the url for embedded instances
func SetServerURL(cmd *cobra.Command, client *kafkarestv3.APIClient, url string) {
	url = strings.TrimSuffix(url, "/")   // localhost:8091/kafka/v3/ --> localhost:8091/kafka/v3
	url = strings.TrimSuffix(url, "/v3") // localhost:8091/kafka/v3 --> localhost:8091/kafka
	protocolRgx := regexp.MustCompile(`(\w+)://`)
	protocolMatch := protocolRgx.MatchString(url)
	if !protocolMatch {
		var protocolMsg string
		if cmd.Flags().Changed("client-cert-path") || cmd.Flags().Changed("certificate-authority-path") { // assume https if client-cert is set since this means we want to use mTLS auth
			url = "https://" + url
			protocolMsg = "Assuming https protocol.\n"
		} else {
			url = "http://" + url
			protocolMsg = "Assuming http protocol.\n"
		}
		if i, _ := cmd.Flags().GetCount("verbose"); i > 0 {
			output.ErrPrint(false, protocolMsg)
		}
	}
	client.ChangeBasePath(strings.TrimSuffix(url, "/") + "/v3")
}

// getKafkaRestUrl tries to fetch the Kafka REST URL from the --url flag or from the CONFLUENT_REST_URL environment variable.
func getKafkaRestUrl(cmd *cobra.Command) (string, error) {
	if url, _ := cmd.Flags().GetString("url"); url != "" {
		return url, nil
	}
	if url := os.Getenv("CONFLUENT_REST_URL"); url != "" {
		return url, nil
	}
	return "", errors.NewErrorWithSuggestions(errors.KafkaRestUrlNotFoundErrorMsg, errors.KafkaRestUrlNotFoundSuggestions)
}
