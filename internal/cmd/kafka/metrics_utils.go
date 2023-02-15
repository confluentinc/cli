package kafka

import (
	"fmt"
	"math"
	"time"

	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	ClusterLoadMetricName                        = "io.confluent.kafka.server/cluster_load_percent"
	threeDayLookbackWindow                       = "P3D/now"
	latestLookbackWindow                         = "PT15M/now"
	hourGranularity        metricsv2.Granularity = "PT1H"
	minuteGranularity      metricsv2.Granularity = "PT1M"
)

func getMetricsOptions(isLatestMetric bool) (metricsv2.Granularity, string, int32) {
	if isLatestMetric {
		// Return latest metric in a 15 minute window
		return minuteGranularity, latestLookbackWindow, 15
	} else {
		// Default to return max metric over a three day window
		return hourGranularity, threeDayLookbackWindow, 1000
	}
}

func getMetricsApiRequest(metricName string, agg string, clusterId string, isLatestMetric bool) metricsv2.QueryRequest {
	granularity, lookback, limit := getMetricsOptions(isLatestMetric)
	aggFunc := metricsv2.AggregationFunction(agg)
	nullableAggFunc := metricsv2.NewNullableAggregationFunction(&aggFunc)
	aggregations := []metricsv2.Aggregation{
		{
			Metric: metricName,
			Agg:    *nullableAggFunc,
		},
	}
	filter := metricsv2.Filter{
		FieldFilter: &metricsv2.FieldFilter{
			Field: metricsv2.PtrString("resource.kafka.id"),
			Op:    "EQ",
			Value: metricsv2.StringAsFieldFilterValue(metricsv2.PtrString(clusterId)),
		},
	}
	req := metricsv2.NewQueryRequest(aggregations, granularity, []string{lookback})
	req.SetFilter(filter)
	req.SetLimit(limit)
	return *req
}

func maxApiDataValue(metricsData []metricsv2.Point) metricsv2.Point {
	maxApiData := metricsv2.Point{
		Value: float32(math.Inf(-1)),
	}
	for _, value := range metricsData {
		if value.Value > maxApiData.Value {
			maxApiData = value
		}
	}
	return maxApiData
}

func (c *clusterCommand) validateClusterLoad(clusterId string, isLatestMetric bool) error {
	query := getMetricsApiRequest(ClusterLoadMetricName, "MAX", clusterId, isLatestMetric)
	clusterLoadResponse, httpResp, err := c.V2Client.MetricsDatasetQuery("cloud", query)
	if err != nil && !ccloudv2.IsDataMatchesMoreThanOneSchemaError(err) || clusterLoadResponse == nil {
		return errors.Errorf("could not retrieve cluster load metrics to validate request to shrink cluster, please try again in a few minutes: %v", err)
	}

	err = ccloudv2.UnmarshalFlatQueryResponseIfDataSchemaMatchError(err, clusterLoadResponse, httpResp)
	if err != nil {
		return err
	}
	maxClusterLoad := maxApiDataValue(clusterLoadResponse.FlatQueryResponse.GetData())
	if maxClusterLoad.Value >= 0.7 {
		return fmt.Errorf("Cluster Load was %f percent at %s.\nRecommended cluster load should be less than 70 percent", maxClusterLoad.Value*100, maxClusterLoad.Timestamp.In(time.Local))
	}
	return nil
}
