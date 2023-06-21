package ccloudv2

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"
)

type flatQueryResponse struct {
	Data []responseDataPoint `json:"data"`
}

type responseDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float32   `json:"value"`
}

type MetricsClient struct {
	*metricsv2.APIClient
	authToken string
}

func NewMetricsClient(url, userAgent string, unsafeTrace bool, authToken string) *MetricsClient {
	cfg := metricsv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = metricsv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return &MetricsClient{
		APIClient: metricsv2.NewAPIClient(cfg),
		authToken: authToken,
	}
}

func (c *MetricsClient) metricsApiContext() context.Context {
	return context.WithValue(context.Background(), metricsv2.ContextAccessToken, c.authToken)
}

func (c *MetricsClient) MetricsDatasetQuery(dataset string, query metricsv2.QueryRequest) (*metricsv2.QueryResponse, *http.Response, error) {
	return c.Version2Api.V2MetricsDatasetQueryPost(c.metricsApiContext(), dataset).QueryRequest(query).Execute()
}

func UnmarshalFlatQueryResponseIfDataSchemaMatchError(err error, metricsResponse *metricsv2.QueryResponse, httpResp *http.Response) error {
	if IsDataMatchesMoreThanOneSchemaError(err) {
		body, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return err
		}
		var resBody flatQueryResponse
		if err := json.Unmarshal(body, &resBody); err != nil {
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
	return err != nil && err.Error() == "Data matches more than one schema in oneOf(QueryResponse)"
}
