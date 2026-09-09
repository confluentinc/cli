package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newKubernetesClusterUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a Kubernetes cluster registered with CMF.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.kubernetesClusterUpdate,
	}

	addCmfFlagSet(cmd)
	cmd.Flags().String("lifecycle-state", "", "Lifecycle state for the Kubernetes cluster (ACTIVE or DECOMMISSIONED).")
	cobra.CheckErr(cmd.MarkFlagRequired("lifecycle-state"))
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) kubernetesClusterUpdate(cmd *cobra.Command, args []string) error {
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	clusterName := args[0]

	lifecycleState, err := cmd.Flags().GetString("lifecycle-state")
	if err != nil {
		return fmt.Errorf("failed to read lifecycle-state: %v", err)
	}

	existingCluster, err := client.DescribeKubernetesCluster(c.createContext(), clusterName)
	if err != nil {
		return err
	}

	existingCluster.Spec = cmfsdk.KubernetesClusterSpec{}
	existingCluster.Spec.SetLifecycleState(lifecycleState)

	updatedCluster, err := client.UpdateKubernetesCluster(c.createContext(), clusterName, existingCluster)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
		o := &kubernetesClusterOutput{
			Name: updatedCluster.Metadata.Name,
		}
		if updatedCluster.Metadata.CreationTimestamp != nil {
			o.CreatedTime = *updatedCluster.Metadata.CreationTimestamp
		}
		if updatedCluster.Metadata.UpdateTimestamp != nil {
			o.UpdatedTime = *updatedCluster.Metadata.UpdateTimestamp
		}
		if updatedCluster.Spec.LifecycleState != nil {
			o.LifecycleState = *updatedCluster.Spec.LifecycleState
		}
		if updatedCluster.Status != nil {
			if updatedCluster.Status.State != nil {
				o.ConnectionState = *updatedCluster.Status.State
			}
			if updatedCluster.Status.KubernetesVersion != nil {
				o.KubernetesVersion = *updatedCluster.Status.KubernetesVersion
			}
		}
		table.Add(o)
		return table.Print()
	}

	localCluster := convertSdkKubernetesClusterToLocal(updatedCluster)
	return output.SerializedOutput(cmd, localCluster)
}
