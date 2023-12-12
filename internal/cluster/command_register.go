package cluster

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	"github.com/confluentinc/cli/v3/pkg/cluster"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
)

type registerCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newRegisterCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register cluster.",
		Long:  "Register cluster with the MDS cluster registry.",
		Args:  cobra.NoArgs,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Register a cluster with the MDS cluster registry:",
				Code: "confluent cluster register --cluster-name myCluster2 --kafka-cluster-id LRx92c9yQ+ws786HYosuBn --hosts 10.6.6.6:9000, 10.3.3.3:9003 --protocol SASL_PLAINTEXT"
			},
		),
	}

	c := &registerCommand{AuthenticatedCLICommand: pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)}
	cmd.RunE = c.register

	cmd.Flags().StringSlice("hosts", []string{}, "A comma-separated list of hosts.")
	cmd.Flags().String("protocol", "", "Security protocol.")
	cmd.Flags().String("cluster-name", "", "Cluster name.")
	cmd.Flags().String("kafka-cluster", "", "Kafka cluster ID.")
	cmd.Flags().String("schema-registry-cluster", "", "Schema Registry cluster ID.")
	cmd.Flags().String("ksql-cluster", "", "ksqlDB cluster ID.")
	cmd.Flags().String("connect-cluster", "", "Kafka Connect cluster ID.")
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("cluster-name"))
	cobra.CheckErr(cmd.MarkFlagRequired("kafka-cluster"))
	cobra.CheckErr(cmd.MarkFlagRequired("hosts"))
	cobra.CheckErr(cmd.MarkFlagRequired("protocol"))

	return cmd
}

func (c *registerCommand) register(cmd *cobra.Command, _ []string) error {
	clusterName, err := cmd.Flags().GetString("cluster-name")
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

	ctx := context.WithValue(context.Background(), mdsv1.ContextAccessToken, c.Context.GetAuthToken())
	clusterInfo := mdsv1.ClusterInfo{ClusterName: clusterName, Scope: mdsv1.Scope{Clusters: *scopeClusters}, Hosts: hosts, Protocol: protocol}

	response, err := c.MDSClient.ClusterRegistryApi.UpdateClusters(ctx, []mdsv1.ClusterInfo{clusterInfo})
	if err != nil {
		return cluster.HandleClusterError(err, response)
	}

	// On Success display the newly added/updated entry
	return cluster.PrintClusters(cmd, []mdsv1.ClusterInfo{clusterInfo})
}

func (c *registerCommand) resolveClusterScope(cmd *cobra.Command) (*mdsv1.ScopeClusters, error) {
	scope := &mdsv1.ScopeClusters{}

	nonKafkaScopesSet := 0

	cmd.Flags().Visit(func(flag *pflag.Flag) {
		switch flag.Name {
		case "kafka-cluster":
			scope.KafkaCluster = flag.Value.String()
		case "schema-registry-cluster":
			scope.SchemaRegistryCluster = flag.Value.String()
			nonKafkaScopesSet++
		case "ksql-cluster":
			scope.KsqlCluster = flag.Value.String()
			nonKafkaScopesSet++
		case "connect-cluster":
			scope.ConnectCluster = flag.Value.String()
			nonKafkaScopesSet++
		}
	})

	if scope.KafkaCluster == "" && nonKafkaScopesSet > 0 {
		return nil, fmt.Errorf(errors.SpecifyKafkaIdErrorMsg)
	}

	if scope.KafkaCluster == "" && nonKafkaScopesSet == 0 {
		return nil, fmt.Errorf("must specify at least one cluster ID")
	}

	if nonKafkaScopesSet > 1 {
		return nil, fmt.Errorf(errors.MoreThanOneNonKafkaErrorMsg)
	}

	return scope, nil
}

func (c *registerCommand) parseHosts(cmd *cobra.Command) ([]mdsv1.HostInfo, error) {
	hosts, err := cmd.Flags().GetStringSlice("hosts")
	if err != nil {
		return nil, err
	}

	hostInfos := make([]mdsv1.HostInfo, len(hosts))
	for i, host := range hosts {
		hostInfo := strings.Split(host, ":")
		port := int64(0)
		if len(hostInfo) > 1 {
			port, _ = strconv.ParseInt(hostInfo[1], 10, 32)
		}
		hostInfos[i] = mdsv1.HostInfo{
			Host: hostInfo[0],
			Port: int32(port),
		}
	}
	return hostInfos, nil
}

func (c *registerCommand) parseProtocol(cmd *cobra.Command) (mdsv1.Protocol, error) {
	protocol, err := cmd.Flags().GetString("protocol")
	if err != nil {
		return "", err
	}

	switch strings.ToUpper(protocol) {
	case "SASL_PLAINTEXT":
		return mdsv1.PROTOCOL_SASL_PLAINTEXT, nil
	case "SASL_SSL":
		return mdsv1.PROTOCOL_SASL_SSL, nil
	case "HTTP":
		return mdsv1.PROTOCOL_HTTP, nil
	case "HTTPS":
		return mdsv1.PROTOCOL_HTTPS, nil
	default:
		return "", fmt.Errorf("protocol %s is currently not supported", protocol)
	}
}
