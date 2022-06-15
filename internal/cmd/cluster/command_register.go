package cluster

import (
	"context"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/spf13/pflag"

	print "github.com/confluentinc/cli/internal/pkg/cluster"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type registerCommand struct {
	*pcmd.AuthenticatedCLICommand
}

const (
	kafkaClusterId   = "kafka-cluster-id"
	srClusterId      = "schema-registry-cluster-id"
	ksqlClusterId    = "ksql-cluster-id"
	connectClusterId = "connect-cluster-id"
)

func newRegisterCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register cluster.",
		Long:  "Register cluster with the MDS cluster registry.",
		Args:  cobra.NoArgs,
	}

	c := &registerCommand{AuthenticatedCLICommand: pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)}
	c.RunE = c.register

	c.Flags().String("hosts", "", "A comma separated list of hosts.")
	c.Flags().String("protocol", "", "Security protocol.")
	c.Flags().String("cluster-name", "", "Cluster name.")
	c.Flags().String("kafka-cluster-id", "", "Kafka cluster ID.")
	c.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID.")
	c.Flags().String("ksql-cluster-id", "", "ksqlDB cluster ID.")
	c.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID.")
	pcmd.AddContextFlag(cmd, c.CLICommand)

	_ = c.MarkFlagRequired("cluster-name")
	_ = c.MarkFlagRequired("kafka-cluster-id")
	_ = c.MarkFlagRequired("hosts")
	_ = c.MarkFlagRequired("protocol")

	return c.Command
}

func (c *registerCommand) register(cmd *cobra.Command, _ []string) error {
	name, err := cmd.Flags().GetString("cluster-name")
	if err != nil {
		return err
	}

	scopeClusters, err := c.resolveClusterScope(cmd)
	if err != nil {
		return err
	}

	hosts, err := c.parseHosts(cmd)
	if err != nil {
		return err
	}

	protocol, err := c.parseProtocol(cmd)
	if err != nil {
		return err
	}

	ctx := context.WithValue(context.Background(), mds.ContextAccessToken, c.State.AuthToken)
	clusterInfo := mds.ClusterInfo{ClusterName: name, Scope: mds.Scope{Clusters: *scopeClusters}, Hosts: hosts, Protocol: protocol}

	response, err := c.MDSClient.ClusterRegistryApi.UpdateClusters(ctx, []mds.ClusterInfo{clusterInfo})
	if err != nil {
		return print.HandleClusterError(err, response)
	}

	// On Success display the newly added/updated entry
	return print.PrintCluster([]mds.ClusterInfo{clusterInfo}, output.Human.String())
}

func (c *registerCommand) resolveClusterScope(cmd *cobra.Command) (*mds.ScopeClusters, error) {
	scope := &mds.ScopeClusters{}

	nonKafkaScopesSet := 0

	cmd.Flags().Visit(func(flag *pflag.Flag) {
		switch flag.Name {
		case kafkaClusterId:
			scope.KafkaCluster = flag.Value.String()
		case srClusterId:
			scope.SchemaRegistryCluster = flag.Value.String()
			nonKafkaScopesSet++
		case ksqlClusterId:
			scope.KsqlCluster = flag.Value.String()
			nonKafkaScopesSet++
		case connectClusterId:
			scope.ConnectCluster = flag.Value.String()
			nonKafkaScopesSet++
		}
	})

	if scope.KafkaCluster == "" && nonKafkaScopesSet > 0 {
		return nil, errors.New(errors.SpecifyKafkaIDErrorMsg)
	}

	if scope.KafkaCluster == "" && nonKafkaScopesSet == 0 {
		return nil, errors.New(errors.MustSpecifyOneClusterIDErrorMsg)
	}

	if nonKafkaScopesSet > 1 {
		return nil, errors.New(errors.MoreThanOneNonKafkaErrorMsg)
	}

	return scope, nil
}

func (c *registerCommand) parseHosts(cmd *cobra.Command) ([]mds.HostInfo, error) {
	hostStr, err := cmd.Flags().GetString("hosts")
	if err != nil {
		return nil, err
	}

	var hostInfos []mds.HostInfo
	for _, host := range strings.Split(hostStr, ",") {
		hostInfo := strings.Split(host, ":")
		port := int64(0)
		if len(hostInfo) > 1 {
			port, _ = strconv.ParseInt(hostInfo[1], 10, 32)
		}
		hostInfos = append(hostInfos, mds.HostInfo{Host: hostInfo[0], Port: int32(port)})
	}
	return hostInfos, nil
}

func (c *registerCommand) parseProtocol(cmd *cobra.Command) (mds.Protocol, error) {
	protocol, err := cmd.Flags().GetString("protocol")
	if err != nil {
		return "", err
	}

	switch strings.ToUpper(protocol) {
	case "SASL_PLAINTEXT":
		return mds.PROTOCOL_SASL_PLAINTEXT, nil
	case "SASL_SSL":
		return mds.PROTOCOL_SASL_SSL, nil
	case "HTTP":
		return mds.PROTOCOL_HTTP, nil
	case "HTTPS":
		return mds.PROTOCOL_HTTPS, nil
	default:
		return "", errors.Errorf(errors.ProtocolNotSupportedErrorMsg, protocol)
	}
}
