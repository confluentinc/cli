package cluster

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

const (
	connectClusterTypeName        = "connect-cluster"
	kafkaClusterTypeName          = "kafka-cluster"
	ksqlClusterTypeName           = "ksql-cluster"
	schemaRegistryClusterTypeName = "schema-registry-cluster"
)

type prettyCluster struct {
	Name           string `human:"Name" serialized:"name"`
	Type           string `human:"Type" serialized:"type"`
	KafkaClusterId string `human:"Kafka Cluster" serialized:"kafka_cluster_id"`
	ComponentId    string `human:"Component ID" serialized:"component_id"`
	Hosts          string `human:"Hosts" serialized:"hosts"`
	Protocol       string `human:"Protocol" serialized:"protocol"`
}

func PrintClusters(cmd *cobra.Command, clusterInfos []mds.ClusterInfo) error {
	if output.GetFormat(cmd).IsSerialized() {
		return output.SerializedOutput(cmd, clusterInfos)
	}

	list := output.NewList(cmd)
	for _, clusterInfo := range clusterInfos {
		list.Add(createPrettyCluster(clusterInfo))
	}
	return list.Print()
}

func createPrettyProtocol(protocol mds.Protocol) string {
	switch protocol {
	case mds.PROTOCOL_SASL_PLAINTEXT, mds.PROTOCOL_SASL_SSL, mds.PROTOCOL_HTTP, mds.PROTOCOL_HTTPS:
		return string(protocol)
	default:
		return ""
	}
}

func createPrettyCluster(clusterInfo mds.ClusterInfo) *prettyCluster {
	var t, id, cid, p string
	switch {
	case clusterInfo.Scope.Clusters.ConnectCluster != "":
		t = connectClusterTypeName
		id = clusterInfo.Scope.Clusters.KafkaCluster
		cid = clusterInfo.Scope.Clusters.ConnectCluster
	case clusterInfo.Scope.Clusters.KsqlCluster != "":
		t = ksqlClusterTypeName
		id = clusterInfo.Scope.Clusters.KafkaCluster
		cid = clusterInfo.Scope.Clusters.KsqlCluster
	case clusterInfo.Scope.Clusters.SchemaRegistryCluster != "":
		t = schemaRegistryClusterTypeName
		id = clusterInfo.Scope.Clusters.KafkaCluster
		cid = clusterInfo.Scope.Clusters.SchemaRegistryCluster
	default:
		t = kafkaClusterTypeName
		cid = ""
		id = clusterInfo.Scope.Clusters.KafkaCluster
	}
	hosts := make([]string, len(clusterInfo.Hosts))
	for i, hostInfo := range clusterInfo.Hosts {
		hosts[i] = createPrettyHost(hostInfo)
	}
	p = createPrettyProtocol(clusterInfo.Protocol)
	return &prettyCluster{
		Name:           clusterInfo.ClusterName,
		Type:           t,
		KafkaClusterId: id,
		ComponentId:    cid,
		Hosts:          strings.Join(hosts, ", "),
		Protocol:       p,
	}
}

func createPrettyHost(hostInfo mds.HostInfo) string {
	if hostInfo.Port > 0 {
		return fmt.Sprintf("%s:%d", hostInfo.Host, hostInfo.Port)
	}
	return hostInfo.Host
}

func HandleClusterError(err error, response *http.Response) error {
	if response != nil && response.StatusCode == http.StatusNotFound {
		return errors.NewWrapErrorWithSuggestions(err, errors.AccessClusterRegistryErrorMsg, errors.AccessClusterRegistrySuggestions)
	}
	return err
}
