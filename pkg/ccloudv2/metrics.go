package ccloudv2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"

	"github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/config"
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

// MetricsDatasetQueryRaw posts a hand-built JSON body to /v2/metrics/{dataset}/query.
// Use this when the request needs a field the typed SDK doesn't expose (e.g. the
// undocumented "time_agg" knob that the Metrics API requires to override the
// gauge MEAN time aggregation; see schema-registry cluster describe).
func (c *MetricsClient) MetricsDatasetQueryRaw(dataset string, body []byte) (*metricsv2.QueryResponse, *http.Response, error) {
	cfg := c.GetConfig()
	if len(cfg.Servers) == 0 {
		return nil, nil, fmt.Errorf("metrics client has no configured server")
	}
	url := fmt.Sprintf("%s/v2/metrics/%s/query", cfg.Servers[0].URL, dataset)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if token, err := auth.GetDataplaneToken(c.cfg.Context()); err == nil {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	httpResp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, httpResp, err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, httpResp, err
	}
	if httpResp.StatusCode >= 400 {
		return nil, httpResp, fmt.Errorf("metrics API returned %d: %s", httpResp.StatusCode, string(respBody))
	}

	var flat flatQueryResponse
	if err := json.Unmarshal(respBody, &flat); err != nil {
		return nil, httpResp, err
	}
	points := make([]metricsv2.Point, len(flat.Data))
	for i, p := range flat.Data {
		points[i] = metricsv2.Point{Value: p.Value, Timestamp: p.Timestamp}
	}
	return &metricsv2.QueryResponse{FlatQueryResponse: metricsv2.NewFlatQueryResponse(points)}, httpResp, nil
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
