package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newKubernetesClusterListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kubernetes clusters registered with CMF.",
		Args:  cobra.NoArgs,
		RunE:  c.kubernetesClusterList,
	}

	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) kubernetesClusterList(cmd *cobra.Command, _ []string) error {
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	clusters, err := client.ListKubernetesClusters(c.createContext())
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		list.Filter([]string{"Name", "CreatedTime", "UpdatedTime", "LifecycleState", "ConnectionState"})
		for _, cluster := range clusters {
			o := &kubernetesClusterOutput{
				Name: cluster.Metadata.Name,
			}
			if cluster.Metadata.CreationTimestamp != nil {
				o.CreatedTime = *cluster.Metadata.CreationTimestamp
			}
			if cluster.Metadata.UpdateTimestamp != nil {
				o.UpdatedTime = *cluster.Metadata.UpdateTimestamp
			}
			if cluster.Spec.LifecycleState != nil {
				o.LifecycleState = *cluster.Spec.LifecycleState
			}
			if cluster.Status != nil && cluster.Status.State != nil {
				o.ConnectionState = *cluster.Status.State
			}
			list.Add(o)
		}
		return list.Print()
	}

	localClusters := make([]LocalKubernetesCluster, 0, len(clusters))
	for _, cluster := range clusters {
		localClusters = append(localClusters, convertSdkKubernetesClusterToLocal(cluster))
	}
	return output.SerializedOutput(cmd, localClusters)
}
