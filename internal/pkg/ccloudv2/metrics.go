package ccloudv2

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"

	testserver "github.com/confluentinc/cli/test/test-server"
)

type flatQueryResponse struct {
	Data []responseDataPoint `json:"data"`
}

type responseDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float32   `json:"value"`
}

func newMetricsClient(baseUrl, userAgent string, unsafeTrace, isTest bool) *metricsv2.APIClient {
	cfg := metricsv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = metricsv2.ServerConfigurations{{URL: getMetricsServerUrl(baseUrl, isTest)}}
	cfg.UserAgent = userAgent

	return metricsv2.NewAPIClient(cfg)
}

func getMetricsServerUrl(baseURL string, isTest bool) string {
	if isTest {
		return testserver.TestV2CloudUrl.String()
	}
	if strings.Contains(baseURL, "devel") {
		return "https://devel-sandbox-api.telemetry.aws.confluent.cloud"
	} else if strings.Contains(baseURL, "stag") {
		return "https://stag-sandbox-api.telemetry.aws.confluent.cloud"
	}
	return "https://api.telemetry.confluent.cloud"
}

func (c *Client) metricsApiContext() context.Context {
	return context.WithValue(context.Background(), metricsv2.ContextAccessToken, c.JwtToken)
}

func (c *Client) MetricsDatasetQuery(dataset string, query metricsv2.QueryRequest) (*metricsv2.QueryResponse, *http.Response, error) {
	req := c.MetricsClient.Version2Api.V2MetricsDatasetQueryPost(c.metricsApiContext(), dataset).QueryRequest(query)
	return c.MetricsClient.Version2Api.V2MetricsDatasetQueryPostExecute(req)
}

func UnmarshalFlatQueryResponseIfDataSchemaMatchError(err error, metricsResponse *metricsv2.QueryResponse, httpResp *http.Response) error {
	if IsDataMatchesMoreThanOneSchemaError(err) {
		body, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return err
		}
		var resBody flatQueryResponse
		err = json.Unmarshal(body, &resBody)
		if err != nil {
			return err
		}

		metricsResponse.FlatQueryResponse = metricsv2.NewFlatQueryResponse([]metricsv2.Point{})

		for _, dataPoint := range resBody.Data {
			metricsResponse.FlatQueryResponse.Data = append(metricsResponse.FlatQueryResponse.Data,
				metricsv2.Point{Value: dataPoint.Value, Timestamp: dataPoint.Timestamp})
		}
	}
	return nil
}

func IsDataMatchesMoreThanOneSchemaError(err error) bool {
	dataSchemaMatchError := "Data matches more than one schema in oneOf(QueryResponse)"
	if err != nil && err.Error() == dataSchemaMatchError {
		return true
	}
	return false
}
