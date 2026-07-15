package flink

import (
	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type kubernetesClusterOutput struct {
	Name              string `human:"Name" serialized:"name"`
	CreatedTime       string `human:"Created Time" serialized:"created_time"`
	UpdatedTime       string `human:"Updated Time" serialized:"updated_time"`
	LifecycleState    string `human:"Lifecycle State,omitempty" serialized:"lifecycle_state,omitempty"`
	ConnectionState   string `human:"Connection State,omitempty" serialized:"connection_state,omitempty"`
	KubernetesVersion string `human:"Kubernetes Version,omitempty" serialized:"kubernetes_version,omitempty"`
}

func (c *command) newKubernetesClusterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "kubernetes-cluster",
		Short:       "Manage Kubernetes clusters registered with CMF.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.AddCommand(c.newKubernetesClusterListCommand())
	cmd.AddCommand(c.newKubernetesClusterDescribeCommand())
	cmd.AddCommand(c.newKubernetesClusterUpdateCommand())

	return cmd
}

func convertSdkKubernetesClusterToLocal(cluster cmfsdk.KubernetesCluster) LocalKubernetesCluster {
	local := LocalKubernetesCluster{
		ApiVersion: cluster.ApiVersion,
		Kind:       cluster.Kind,
		Metadata: LocalKubernetesClusterMetadata{
			Name:              cluster.Metadata.Name,
			CreationTimestamp: cluster.Metadata.CreationTimestamp,
			UpdateTimestamp:   cluster.Metadata.UpdateTimestamp,
			Uid:               cluster.Metadata.Uid,
			Labels:            cluster.Metadata.Labels,
			Annotations:       cluster.Metadata.Annotations,
		},
		Spec: LocalKubernetesClusterSpec{
			LifecycleState: cluster.Spec.LifecycleState,
		},
	}

	if cluster.Status != nil {
		local.Status = &LocalKubernetesClusterStatus{
			State:                  cluster.Status.State,
			Message:                cluster.Status.Message,
			LastHeartbeatTimestamp: cluster.Status.LastHeartbeatTimestamp,
			KubernetesVersion:      cluster.Status.KubernetesVersion,
		}
	}

	return local
}
