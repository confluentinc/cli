package kafka

import (
	"context"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func initKafkaRest(c *pcmd.AuthenticatedCLICommand, cmd *cobra.Command) (*kafkarestv3.APIClient, context.Context, error) {
	url, err := getKafkaRestUrl(cmd)
	if err != nil { // require the flag
		return nil, nil, err
	}
	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return nil, nil, err
	}
	kafkaRestClient := kafkaREST.Client
	setServerURL(cmd, kafkaRestClient, url)
	return kafkaRestClient, kafkaREST.Context, nil
}

// Used for on-prem KafkaRest commands
// Embedded KafkaRest uses /kafka/v3 and standalone uses /v3
// Relying on users to include the /kafka in the url for embedded instances
func setServerURL(cmd *cobra.Command, client *kafkarestv3.APIClient, url string) {
	url = strings.Trim(url, "/")   // localhost:8091/kafka/v3/ --> localhost:8091/kafka/v3
	url = strings.Trim(url, "/v3") // localhost:8091/kafka/v3 --> localhost:8091/kafka
	protocolRgx, _ := regexp.Compile(`(\w+)://`)
	protocolMatch := protocolRgx.MatchString(url)
	if !protocolMatch {
		var protocolMsg string
		if cmd.Flags().Changed("client-cert-path") || cmd.Flags().Changed("ca-cert-path") { // assume https if client-cert is set since this means we want to use mTLS auth
			url = "https://" + url
			protocolMsg = errors.AssumingHttpsProtocol
		} else {
			url = "http://" + url
			protocolMsg = errors.AssumingHttpProtocol
		}
		if i, _ := cmd.Flags().GetCount("verbose"); i > 0 {
			output.ErrPrintf(protocolMsg)
		}
	}
	client.ChangeBasePath(strings.Trim(url, "/") + "/v3")
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
