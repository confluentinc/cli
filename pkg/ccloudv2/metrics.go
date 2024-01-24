package ccloudv2

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"

	"github.com/confluentinc/cli/v3/pkg/auth"
	"github.com/confluentinc/cli/v3/pkg/config"
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
	cfg *config.Config
}

func NewMetricsClient(configuration *metricsv2.Configuration, cfg *config.Config) *MetricsClient {
	return &MetricsClient{
		APIClient: metricsv2.NewAPIClient(configuration),
		cfg:       cfg,
	}
}

func (c *MetricsClient) context() context.Context {
	ctx := context.Background()

	dataplaneToken, err := auth.GetDataplaneToken(c.cfg.Context())
	if err != nil {
		return ctx
	}

	return context.WithValue(ctx, metricsv2.ContextAccessToken, dataplaneToken)
}

func (c *MetricsClient) MetricsDatasetQuery(dataset string, query metricsv2.QueryRequest) (*metricsv2.QueryResponse, *http.Response, error) {
	return c.Version2Api.V2MetricsDatasetQueryPost(c.context(), dataset).QueryRequest(query).Execute()
}

func UnmarshalFlatQueryResponseIfDataSchemaMatchError(err error, metricsResponse *metricsv2.QueryResponse, httpResp *http.Response) error {
	if !IsDataMatchesMoreThanOneSchemaError(err) {
		return nil
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	var resBody flatQueryResponse
	if err := json.Unmarshal(body, &resBody); err != nil {
		return err
	}

	points := make([]metricsv2.Point, len(resBody.Data))
	for i, dataPoint := range resBody.Data {
		points[i] = metricsv2.Point{
			Value:     dataPoint.Value,
			Timestamp: dataPoint.Timestamp,
		}
	}

	metricsResponse.FlatQueryResponse = metricsv2.NewFlatQueryResponse(points)
	return nil
}

func IsDataMatchesMoreThanOneSchemaError(err error) bool {
	return err != nil && err.Error() == "Data matches more than one schema in oneOf(QueryResponse)"
}
