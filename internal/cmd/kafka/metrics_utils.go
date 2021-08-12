package kafka

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/confluentinc/ccloud-sdk-go-v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	ClusterLoadMetricName  = "io.confluent.kafka.server/cluster_load_percent"
	PartitionMetricName    = "io.confluent.kafka.server/partition_count"
	StorageMetricName      = "io.confluent.kafka.server/retained_bytes"
	latestLookbackWindow   = "PT15M"
	threeDayLookbackWindow = "P3D"
	hourGranularity        = "PT1H"
	minuteGranularity      = "PT1M"
)

func getMetricsOptions(isLatestMetric bool) (string, string, int32) {
	if isLatestMetric {
		// Return latest metric in a 15 minute window
		return minuteGranularity, latestLookbackWindow, 15
	} else {
		// Default to return max metric over a three day window
		return hourGranularity, threeDayLookbackWindow, 1000
	}
}
func getMetricsApiRequest(metricName string, agg string, clusterId string, isLatestMetric bool) *ccloud.MetricsApiRequest {
	granularity, lookback, limit := getMetricsOptions(isLatestMetric)
	return &ccloud.MetricsApiRequest{
		Aggregations: []ccloud.ApiAggregation{
			{
				Metric: metricName,
				Agg:    agg,
			},
		},
		Filter: ccloud.ApiFilter{
			Field: "resource.kafka.id",
			Op:    "EQ",
			Value: clusterId,
		},
		Granularity: granularity,
		Lookback:    lookback,
		Limit:       limit,
	}
}

func maxApiDataValue(metricsData []ccloud.ApiData) ccloud.ApiData {
	max := metricsData[0]
	for _, value := range metricsData {
		if value.Value > max.Value {
			max = value
		}
	}
	return max
}

func (c *clusterCommand) validateClusterLoad(clusterId string, isLatestMetric bool) error {
	// Get Cluster Load Metric
	clusterLoadResponse, err := c.Client.MetricsApi.QueryV2(
		context.Background(), "cloud", getMetricsApiRequest(ClusterLoadMetricName, "MAX", clusterId, isLatestMetric), "")
	if err != nil || clusterLoadResponse == nil || len(clusterLoadResponse.Result) == 0 {
		c.logger.Warn("Could not retrieve Cluster Load Metrics: ", err)
		return errors.Errorf("Could not retrieve cluster load metrics to validate request to shrink cluster. Please try again in a few minutes. err: %v", err)
	}
	maxClusterLoad := maxApiDataValue(clusterLoadResponse.Result)
	if maxClusterLoad.Value > 0.7 {
		return fmt.Errorf("\nCluster Load was %f percent at %s. \nRecommended cluster load should be less than 70 percent", maxClusterLoad.Value*100, maxClusterLoad.Timestamp.In(time.Local))
	}
	return nil
}

func (c *clusterCommand) validatePartitionCount(clusterId string, requiredPartitionCount int32, isLatestMetric bool, cku int32) error {
	partitionMetricsResponse, err := c.Client.MetricsApi.QueryV2(
		context.Background(), "cloud", getMetricsApiRequest(PartitionMetricName, "SUM", clusterId, isLatestMetric), "")
	if err != nil || partitionMetricsResponse == nil || len(partitionMetricsResponse.Result) == 0 {
		return errors.Errorf("Could not retrieve partition count metrics to validate request to shrink cluster. Please try again in a few minutes. err: %v", err)
	}

	maxPartitionCount := maxApiDataValue(partitionMetricsResponse.Result)
	if int32(maxPartitionCount.Value) > requiredPartitionCount {
		return fmt.Errorf("partition count was %f at %s.\nRecommended partition count is less than %d for %d cku", maxPartitionCount.Value, maxPartitionCount.Timestamp.In(time.Local), requiredPartitionCount, cku)
	}
	return nil
}

func (c *clusterCommand) validateStorageLimit(clusterId string, requiredStorageLimit int32, isLatestMetric bool, cku int32) error {
	storageMetricsResponse, err := c.Client.MetricsApi.QueryV2(
		context.Background(), "cloud", getMetricsApiRequest(StorageMetricName, "SUM", clusterId, isLatestMetric), "")
	if err != nil || storageMetricsResponse == nil || len(storageMetricsResponse.Result) == 0 {
		return errors.Errorf("Could not retrieve storage metrics to validate request to shrink cluster. Please try again in a few minutes. err: %v", err)
	}
	maxStorageLimit := maxApiDataValue(storageMetricsResponse.Result)
	maxStorageLimitInGB := maxStorageLimit.Value * math.Pow10(-9)
	if maxStorageLimitInGB > float64(requiredStorageLimit) {
		return fmt.Errorf("storage used was %.5f at %s. Recommended storage is less than %d for %d CKU", maxStorageLimitInGB, maxStorageLimit.Timestamp.In(time.Local), requiredStorageLimit, cku)
	}
	return nil
}
