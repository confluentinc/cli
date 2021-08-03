package kafka

import (
	"context"
	"fmt"
	"math"

	"github.com/confluentinc/ccloud-sdk-go-v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	ClusterLoadMetricName  = "io.confluent.kafka.server/cluster_load_percent"
	PartitionMetricName    = "io.confluent.kafka.server/partition_count"
	StorageMetricName      = "io.confluent.kafka.server/retained_bytes"
	latestLookbackWindow   = "PT15M"
	threeDayLookbackWindow = "PT3D"
	hourGranularity        = "PT1H"
	minuteGranularity      = "PT1M"
)

func getMetricsOptions(isLatestMetric bool) (string, string, int32) {
	if isLatestMetric {
		// Return latest metric in a 15 minute window
		return minuteGranularity, latestLookbackWindow, 1
	} else {
		// Default to return max metric over a three day window
		return hourGranularity, threeDayLookbackWindow, 1000
	}
}
func getMetricsApiRequest(metricName string, clusterId string, isLatestMetric bool) *ccloud.MetricsApiRequest {
	granularity, lookback, limit := getMetricsOptions(isLatestMetric)
	return &ccloud.MetricsApiRequest{
		Aggregations: []ccloud.ApiAggregation{
			{
				Metric: metricName,
				Agg:    "SUM",
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
		context.Background(), "cloud", getMetricsApiRequest(ClusterLoadMetricName, clusterId, isLatestMetric), "")
	if err != nil || clusterLoadResponse == nil || len(clusterLoadResponse.Result) == 0 {
		c.logger.Warn("Could not retrieve Cluster Load Metrics: ", err)
		return errors.New("Could not retrieve cluster load metrics to validate request to shrink cluster. Please try again in a few minutes.")
	}
	if isLatestMetric {
		latestClusterLoad := clusterLoadResponse.Result[len(clusterLoadResponse.Result)-1].Value
		timestamp := clusterLoadResponse.Result[len(clusterLoadResponse.Result)-1].Timestamp
		if latestClusterLoad > 0.7 {
			return fmt.Errorf("\nCluster Load was %f percent at %s in the last 15 mins.\nRecommended cluster load should be less than 70 percent", latestClusterLoad*100, timestamp.Format("2006-01-02 15:04:05"))
		}
	} else {
		maxClusterLoad := maxApiDataValue(clusterLoadResponse.Result)
		if maxClusterLoad.Value > 0.7 {
			return fmt.Errorf("\nCluster Load was %f percent at %s in the last 3 days. \nRecommended cluster load should be less than 70 percent", maxClusterLoad.Value*100, maxClusterLoad.Timestamp.Format("2006-01-02 15:04:05"))
		}
	}
	return nil
}

func (c *clusterCommand) validatePartitionCount(clusterId string, requiredPartitionCount int32, isLatestMetric bool, cku int32) error {
	partitionMetricsResponse, err := c.Client.MetricsApi.QueryV2(
		context.Background(), "cloud", getMetricsApiRequest(PartitionMetricName, clusterId, isLatestMetric), "")
	if err != nil || partitionMetricsResponse == nil || len(partitionMetricsResponse.Result) == 0 {
		return errors.New("Could not retrieve partition count metrics to validate request to shrink cluster. Please try again in a few minutes.")
	}
	if isLatestMetric {
		latestPartitionCount := int32(math.Round(partitionMetricsResponse.Result[len(partitionMetricsResponse.Result)-1].Value))
		timestamp := partitionMetricsResponse.Result[len(partitionMetricsResponse.Result)-1].Timestamp
		if latestPartitionCount > requiredPartitionCount {
			return fmt.Errorf("partition count was %d at %s in the last 15 min.\nRecommended partition count is less than %d for %d cku", latestPartitionCount, timestamp.Format("2006-01-02 15:04:05"), requiredPartitionCount, cku)
		}
	} else {
		maxPartitionCount := maxApiDataValue(partitionMetricsResponse.Result)
		if int32(maxPartitionCount.Value) > requiredPartitionCount {
			return fmt.Errorf("partition count was %f at %s in the last 3 days.\nRecommended partition count is less than %d for %d cku", maxPartitionCount.Value, maxPartitionCount.Timestamp.Format("2006-01-02 15:04:05"), requiredPartitionCount, cku)
		}
	}
	return nil
}

func (c *clusterCommand) validateStorageLimit(clusterId string, requiredStorageLimit int32, isLatestMetric bool, cku int32) error {
	storageMetricsResponse, err := c.Client.MetricsApi.QueryV2(
		context.Background(), "cloud", getMetricsApiRequest(StorageMetricName, clusterId, isLatestMetric), "")
	if err != nil || storageMetricsResponse == nil || len(storageMetricsResponse.Result) == 0 {
		return errors.New("Could not retrieve storage metrics to validate request to shrink cluster. Please try again in a few minutes.")
	}
	if isLatestMetric {
		latestStorageBytes := int32(math.Round(storageMetricsResponse.Result[len(storageMetricsResponse.Result)-1].Value))
		timestamp := storageMetricsResponse.Result[len(storageMetricsResponse.Result)-1].Timestamp
		if latestStorageBytes > requiredStorageLimit {
			return fmt.Errorf("storage used was %d at %s in the last 15 minutes. Recommended storage is less than %d for %d CKU", latestStorageBytes, timestamp.Format("2006-01-02 15:04:05"), requiredStorageLimit, cku)
		}
	} else {
		maxStorageLimit := maxApiDataValue(storageMetricsResponse.Result)
		if int32(maxStorageLimit.Value) > requiredStorageLimit {
			return fmt.Errorf("storage used was %f at %s in the last 3 days. Recommended storage is less than %d for %d CKU", maxStorageLimit.Value, maxStorageLimit.Timestamp.Format("2006-01-02 15:04:05"), requiredStorageLimit, cku)
		}
	}
	return nil
}
