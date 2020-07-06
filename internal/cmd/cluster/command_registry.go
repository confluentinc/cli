package cluster

import (
	"context"
	"fmt"
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

type registryCommand struct {
	*pcmd.AuthenticatedCLICommand
}

const (
	KafkaClusterId   = "kafka-cluster-id"
	SrClusterId      = "schema-registry-cluster-id"
	KSQLClusterId    = "ksql-cluster-id"
	ConnectClusterId = "connect-cluster-id"
)

// NewRegisterCommand registers a cluster to the Cluster Registry in MDS
func NewRegisterCommand(prerunner pcmd.PreRunner) *cobra.Command {
	registerCmd := &registryCommand{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedWithMDSCLICommand(
			&cobra.Command{
				Use:   "register",
				Short: "Register cluster.",
				Long:  "Register cluster with the MDS cluster registry.",
				Args:  cobra.NoArgs,
			},
			prerunner),
	}
	registerCmd.Flags().String("cluster-name", "", "Cluster name.")
	check(registerCmd.MarkFlagRequired("cluster-name"))
	registerCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID.")
	check(registerCmd.MarkFlagRequired("kafka-cluster-id"))
	registerCmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID.")
	registerCmd.Flags().String("ksql-cluster-id", "", "KSQL cluster ID.")
	registerCmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID.")
	registerCmd.Flags().String("hosts", "", "A comma separated list of hosts.")
	check(registerCmd.MarkFlagRequired("hosts"))
	registerCmd.Flags().String("protocol", "", "Security protocol.")
	check(registerCmd.MarkFlagRequired("protocol"))
	registerCmd.Flags().SortFlags = false
	registerCmd.RunE = registerCmd.register
	return registerCmd.Command
}

func NewUnregisterCommand(prerunner pcmd.PreRunner) *cobra.Command {
	unregisterCmd := &registryCommand{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedWithMDSCLICommand(
			&cobra.Command{
				Use:   "unregister",
				Short: "Unregister cluster.",
				Long:  "Unregister cluster from the MDS cluster registry.",
				Args:  cobra.NoArgs,
			},
			prerunner),
	}
	unregisterCmd.Flags().String("cluster-name", "", "Cluster Name.")
	check(unregisterCmd.MarkFlagRequired("cluster-name"))
	unregisterCmd.RunE = unregisterCmd.unregister
	unregisterCmd.Flags().SortFlags = false
	return unregisterCmd.Command
}

func (c *registryCommand) createContext() context.Context {
	return context.WithValue(context.Background(), mds.ContextAccessToken, c.State.AuthToken)
}

func (c *registryCommand) resolveClusterScope(cmd *cobra.Command) (*mds.ScopeClusters, error) {
	scope := &mds.ScopeClusters{}

	nonKafkaScopesSet := 0

	cmd.Flags().Visit(func(flag *pflag.Flag) {
		switch flag.Name {
		case KafkaClusterId:
			scope.KafkaCluster = flag.Value.String()
		case SrClusterId:
			scope.SchemaRegistryCluster = flag.Value.String()
			nonKafkaScopesSet++
		case KSQLClusterId:
			scope.KsqlCluster = flag.Value.String()
			nonKafkaScopesSet++
		case ConnectClusterId:
			scope.ConnectCluster = flag.Value.String()
			nonKafkaScopesSet++
		}
	})

	if scope.KafkaCluster == "" && nonKafkaScopesSet > 0 {
		return nil, errors.New("Must also specify a --kafka-cluster-id to uniquely identify the scope.")
	}

	if scope.KafkaCluster == "" && nonKafkaScopesSet == 0 {
		return nil, errors.New("Must specify at least one cluster ID ")
	}

	if nonKafkaScopesSet > 1 {
		return nil, errors.New("Cannot specify more than one non-Kafka cluster ID for a scope.")
	}

	return scope, nil
}

func (c *registryCommand) parseHosts(cmd *cobra.Command) ([]mds.HostInfo, error) {
	hostStr, err := cmd.Flags().GetString("hosts")
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}

	hostInfos := make([]mds.HostInfo, 0)
	for _, host := range strings.Split(hostStr, ",") {
		hostInfo := strings.Split(host, ":")
		port := 0
		if len(hostInfo) > 1 {
			port, _ = strconv.Atoi(hostInfo[1])
		}
		hostInfos = append(hostInfos, mds.HostInfo{Host: hostInfo[0], Port: int32(port)})
	}
	return hostInfos, nil
}

func (c *registryCommand) parseProtocol(cmd *cobra.Command) (mds.Protocol, error) {
	protocol, err := cmd.Flags().GetString("protocol")
	if err != nil {
		return "", errors.HandleCommon(err, cmd)
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
		return "", fmt.Errorf("Protocol %s is currently not supported.", protocol)
	}
}

func (c *registryCommand) register(cmd *cobra.Command, _ []string) error {

	name, err := cmd.Flags().GetString("cluster-name")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	scopeClusters, err := c.resolveClusterScope(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	hosts, err := c.parseHosts(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	protocol, err := c.parseProtocol(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	clusterInfo := mds.ClusterInfo{ClusterName: name, Scope: mds.Scope{Clusters: *scopeClusters}, Hosts: hosts, Protocol: protocol}

	response, err := c.MDSClient.ClusterRegistryApi.UpdateClusters(c.createContext(), []mds.ClusterInfo{clusterInfo})
	if err != nil {
		return print.HandleClusterError(cmd, err, response)
	}

	// On Success display the newly added/updated entry
	return print.PrintCluster(cmd, []mds.ClusterInfo{clusterInfo}, output.Human.String())
}

func (c *registryCommand) unregister(cmd *cobra.Command, _ []string) error {
	name, err := cmd.Flags().GetString("cluster-name")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	response, err := c.MDSClient.ClusterRegistryApi.DeleteNamedCluster(c.createContext(), name)
	if err != nil {
		return print.HandleClusterError(cmd, err, response)
	}

	pcmd.Printf(cmd, "Successfully unregistered the cluster %s from the Cluster Registry. \n", name)
	return nil
}
