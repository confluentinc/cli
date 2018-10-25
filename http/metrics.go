package http

import (
	metricsv1 "github.com/confluentinc/cc-structs/kafka/metrics/v1"
	"github.com/confluentinc/cli/log"
	"github.com/dghubble/sling"
	"net/http"
	"strings"
)

const (
	kafkaMetricsPath = "/api/kafka_metrics"
)

// MetricsService provides methods for retrieving metrics for Kafka and other components
type MetricsService struct {
	client *http.Client
	sling  *sling.Sling
	logger *log.Logger
}

var _ Metrics = (*MetricsService)(nil)

// NewMetricsService returns a new MetricsService.
func NewMetricsService(client *Client) *MetricsService {
	return &MetricsService{
		client: client.httpClient,
		logger: client.logger,
		sling:  client.sling,
	}
}

// KafkaMetrics returns the Kafka metrics
func (m *MetricsService) KafkaMetrics(logicalClusterIDs []string, dateStart string, dateEnd string) (map[string]*metricsv1.KafkaMetric, *http.Response, error) {
	path := kafkaMetricsPath + "?" + "ids=" + strings.Join(logicalClusterIDs, ",") + "&from=" + dateStart
	if len(dateEnd) != 0 {
		path = path + "to=" + dateEnd
	}
	metricsReply := new(metricsv1.GetKafkaMetricsReply)
	resp, err := m.sling.New().Get(path).Receive(metricsReply, metricsReply)
	if err != nil {
		return nil, resp, err
	}
	if metricsReply.Error != nil {
		return metricsReply.Metrics, resp, metricsReply.Error
	}
	return metricsReply.Metrics, resp, nil
}

// SchemaRegistryMetrics returns Schema Registry metrics
func (m *MetricsService) SchemaRegistryMetrics(logicalClusterID string) (*metricsv1.SchemaRegistryMetric, *http.Response, error) {
	path := "/schema_registries/" + logicalClusterID + "/metrics"
	metricsReply := new(metricsv1.GetSchemaRegistryMetricReply)
	resp, err := m.sling.New().Get(path).Receive(metricsReply, metricsReply)
	if err != nil {
		return nil, resp, err
	}
	if metricsReply.Error != nil {
		return metricsReply.Metric, resp, metricsReply.Error

	}
	return metricsReply.Metric, resp, nil
}
