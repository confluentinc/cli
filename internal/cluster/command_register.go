package cluster

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	"github.com/confluentinc/cli/v4/pkg/cluster"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
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
				Text: "Register a new Confluent Platform cluster:",
				Code: "confluent cluster register --cluster-name myKafkaCluster --kafka-cluster kafka-ID --hosts 10.6.6.6:9000,10.3.3.3:9003 --protocol SASL_PLAINTEXT",
			},
			examples.Example{
				Text: "For more information, see https://docs.confluent.io/platform/current/security/cluster-registry.html#registering-clusters.",
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
	cmd.Flags().String("cmf", "", "Confluent Managed Flink (CMF) ID.")
	cmd.Flags().String("flink-environment", "", "Flink environment ID.")
	cmd.Flags().AddFlagSet(pcmd.OnPremMTLSSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("cluster-name"))
	cobra.CheckErr(cmd.MarkFlagRequired("hosts"))
	cobra.CheckErr(cmd.MarkFlagRequired("protocol"))

	cmd.MarkFlagsRequiredTogether("client-cert-path", "client-key-path")

	return cmd
}

func (c *registerCommand) register(cmd *cobra.Command, _ []string) error {
	client, err := c.GetMDSClient(cmd)
	if err != nil {
		return err
	}

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

	response, err := client.ClusterRegistryApi.UpdateClusters(ctx, []mdsv1.ClusterInfo{clusterInfo})
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
		case "cmf":
			scope.Cmf = flag.Value.String()
		case "flink-environment":
			scope.FlinkEnvironment = flag.Value.String()
		}
	})

	if scope.KafkaCluster == "" && nonKafkaScopesSet > 0 {
		return nil, fmt.Errorf(errors.SpecifyKafkaIdErrorMsg)
	}

	if scope.Cmf == "" && scope.FlinkEnvironment != "" {
		return nil, fmt.Errorf(errors.SpecifyCmfErrorMsg)
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
