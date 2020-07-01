package cluster

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/go-printer"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/olekukonko/tablewriter"
)

type prettyCluster struct {
	Name     string
	Type     string
	ID       string
	CID      string
	Hosts    string
	Protocol string
}

const (
	connectClusterTypeName        = "connect-cluster"
	kafkaClusterTypeName          = "kafka-cluster"
	ksqlClusterTypeName           = "ksql-cluster"
	schemaRegistryClusterTypeName = "schema-registry-cluster"
)

var (
	clusterFields = []string{"Name", "Type", "ID", "CID", "Hosts", "Protocol"}
	clusterLabels = []string{"Name", "Type", " Kafka ID", "Component ID", "Hosts", "Protocol"}
)

func PrintCluster(cmd *cobra.Command, clusterInfos []mds.ClusterInfo, format string) error {
	if format == output.Human.String() {
		var data [][]string
		for _, clusterInfo := range clusterInfos {
			clusterDisplay, err := createPrettyCluster(clusterInfo)
			if err != nil {
				return errors.HandleCommon(err, cmd)
			}
			data = append(data, printer.ToRow(clusterDisplay, clusterFields))
		}
		outputTable(data)
	} else {
		return output.StructuredOutput(format, clusterInfos)
	}

	return nil
}

func createPrettyProtocol(protocol mds.Protocol) string {

	switch protocol {
	case mds.PROTOCOL_SASL_PLAINTEXT:
		return "SASL_PLAINTEXT"
	case mds.PROTOCOL_SASL_SSL:
		return "SASL_SSL"
	case mds.PROTOCOL_HTTP:
		return "HTTP"
	case mds.PROTOCOL_HTTPS:
		return "HTTPS"
	default:
		return ""
	}
}

func createPrettyCluster(clusterInfo mds.ClusterInfo) (*prettyCluster, error) {
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
		hosts[i], _ = createPrettyHost(hostInfo)
	}
	p = createPrettyProtocol(clusterInfo.Protocol)
	return &prettyCluster{
		clusterInfo.ClusterName,
		t,
		id,
		cid,
		strings.Join(hosts, ", "),
		p,
	}, nil
}

func createPrettyHost(hostInfo mds.HostInfo) (string, error) {
	if hostInfo.Port > 0 {
		return fmt.Sprintf("%s:%d", hostInfo.Host, hostInfo.Port), nil
	}
	return hostInfo.Host, nil
}

func outputTable(data [][]string) {
	tablePrinter := tablewriter.NewWriter(os.Stdout)
	tablePrinter.SetAutoWrapText(false)
	tablePrinter.SetAutoFormatHeaders(false)
	tablePrinter.SetHeader(clusterLabels)
	tablePrinter.AppendBulk(data)
	tablePrinter.SetBorder(false)
	tablePrinter.Render()
}

func HandleClusterError(cmd *cobra.Command, err error, response *http.Response) error {
	if response != nil && response.StatusCode == http.StatusNotFound {
		cmd.SilenceUsage = true
		return fmt.Errorf("Unable to access Cluster Registry (%s). Ensure that you're running against MDS with CP 6.0+.", err.Error())
	}
	return errors.HandleCommon(err, cmd)
}
