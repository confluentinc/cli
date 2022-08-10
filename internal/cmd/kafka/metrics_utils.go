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
	PartitionMetricName                          = "io.confluent.kafka.server/partition_count"
	StorageMetricName                            = "io.confluent.kafka.server/retained_bytes"
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
	if maxClusterLoad.Value > 0.7 {
		return fmt.Errorf("\nCluster Load was %f percent at %s. \nRecommended cluster load should be less than 70 percent", maxClusterLoad.Value*100, maxClusterLoad.Timestamp.In(time.Local))
	}
	return nil
}

func (c *clusterCommand) validatePartitionCount(clusterId string, requiredPartitionCount int32, isLatestMetric bool, cku int32) error {
	query := getMetricsApiRequest(PartitionMetricName, "SUM", clusterId, isLatestMetric)
	partitionMetricsResponse, httpResp, err := c.V2Client.MetricsDatasetQuery("cloud", query)
	if err != nil && !ccloudv2.IsDataMatchesMoreThanOneSchemaError(err) || partitionMetricsResponse == nil {
		return errors.Errorf("could not retrieve partition count metrics to validate request to shrink cluster, please try again in a few minutes: %v", err)
	}

	err = ccloudv2.UnmarshalFlatQueryResponseIfDataSchemaMatchError(err, partitionMetricsResponse, httpResp)
	if err != nil {
		return err
	}
	maxPartitionCount := maxApiDataValue(partitionMetricsResponse.FlatQueryResponse.GetData())
	if int32(maxPartitionCount.Value) > requiredPartitionCount {
		return fmt.Errorf("partition count was %f at %s.\nRecommended partition count is less than %d for %d cku", maxPartitionCount.Value, maxPartitionCount.Timestamp.In(time.Local), requiredPartitionCount, cku)
	}
	return nil
}

func (c *clusterCommand) validateStorageLimit(clusterId string, requiredStorageLimit int32, isLatestMetric bool, cku int32) error {
	query := getMetricsApiRequest(StorageMetricName, "SUM", clusterId, isLatestMetric)
	storageMetricsResponse, httpResp, err := c.V2Client.MetricsDatasetQuery("cloud", query)
	if err != nil && !ccloudv2.IsDataMatchesMoreThanOneSchemaError(err) || storageMetricsResponse == nil {
		return errors.Errorf("could not retrieve storage metrics to validate request to shrink cluster, please try again in a few minutes: %v", err)
	}

	err = ccloudv2.UnmarshalFlatQueryResponseIfDataSchemaMatchError(err, storageMetricsResponse, httpResp)
	if err != nil {
		return err
	}
	maxStorageLimit := maxApiDataValue(storageMetricsResponse.FlatQueryResponse.GetData())
	maxStorageLimitInGB := float64(maxStorageLimit.Value) * math.Pow10(-9)
	if maxStorageLimitInGB > float64(requiredStorageLimit) {
		return fmt.Errorf("storage used was %.5f at %s. Recommended storage is less than %d for %d CKU", maxStorageLimitInGB, maxStorageLimit.Timestamp.In(time.Local), requiredStorageLimit, cku)
	}
	return nil
}
