package kafka

import (
	"time"

	"github.com/confluentinc/ccloud-sdk-go-v1"
)

var (
	queryTimeoutDefault          = 180 * time.Second
	queryRetryMinWaitTimeDefault = 100 * time.Millisecond
	queryRetryMaxWaitTimeDefault = 1 * time.Second
	queryNumRetriesDefault       = 2
	latestLookbackWindow         = "PT15M"
	threeDayLookbackWindow       = "PT3D"
	hourGranularity              = "PT1H"
	minuteGranularity            = "PT1M"
)

func getMetricsOptions(isLatestMetric bool) (string, string) {
	if isLatestMetric {
		// Return latest metric in a 15 minute window
		return minuteGranularity, latestLookbackWindow
	} else {
		// Default to return max metric over a three day window
		return hourGranularity, threeDayLookbackWindow
	}
}

func clusterLoadMetricQuery(clusterId string, isLatestMetric bool) *ccloud.MetricsApiRequest {
	granularity, lookback := getMetricsOptions(isLatestMetric)
	return &ccloud.MetricsApiRequest{
		Aggregations: []ccloud.ApiAggregation{
			{
				Metric: "io.confluent.kafka.server/broker_load/cluster_load_percent",
			},
		},
		Filter: ccloud.ApiFilter{
			Field: "resource.kafka.id",
			Op:    "EQ",
			Value: clusterId,
		},
		Granularity: granularity,
		Lookback:    lookback,
	}
}

func partitionCountMetricQuery(clusterId string, isLatestMetric bool) *ccloud.MetricsApiRequest {
	granularity, lookback := getMetricsOptions(isLatestMetric)
	return &ccloud.MetricsApiRequest{
		Aggregations: []ccloud.ApiAggregation{
			{
				Metric: "io.confluent.kafka.server/partition_count",
			},
		},
		Filter: ccloud.ApiFilter{
			Field: "resource.kafka.id",
			Op:    "EQ",
			Value: clusterId,
		},
		Granularity: granularity,
		Lookback:    lookback,
	}
}

func storageBytesMetricQuery(clusterId string, isLatestMetric bool) *ccloud.MetricsApiRequest {
	granularity, lookback := getMetricsOptions(isLatestMetric)
	return &ccloud.MetricsApiRequest{
		Aggregations: []ccloud.ApiAggregation{
			{
				Metric: "io.confluent.kafka.server/retained_bytes",
			},
		},
		Filter: ccloud.ApiFilter{
			Field: "resource.kafka.id",
			Op:    "EQ",
			Value: clusterId,
		},
		Granularity: granularity,
		Lookback:    lookback,
	}
}
