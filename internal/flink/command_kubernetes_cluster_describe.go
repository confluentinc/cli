package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newKubernetesClusterDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Kubernetes cluster registered with CMF.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.kubernetesClusterDescribe,
	}

	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) kubernetesClusterDescribe(cmd *cobra.Command, args []string) error {
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	clusterName := args[0]
	cluster, err := client.DescribeKubernetesCluster(c.createContext(), clusterName)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
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
		if cluster.Status != nil {
			if cluster.Status.State != nil {
				o.ConnectionState = *cluster.Status.State
			}
			if cluster.Status.KubernetesVersion != nil {
				o.KubernetesVersion = *cluster.Status.KubernetesVersion
			}
		}
		table.Add(o)
		return table.Print()
	}

	localCluster := convertSdkKubernetesClusterToLocal(cluster)
	return output.SerializedOutput(cmd, localCluster)
}
