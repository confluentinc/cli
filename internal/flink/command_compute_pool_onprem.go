package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/flink"
)

type computePoolOutOnPrem struct {
	CreationTime string `human:"Creation Time" serialized:"creation_time"`
	Name         string `human:"Name" serialized:"name"`
	Type         string `human:"Type" serialized:"type"`
	Phase        string `human:"Phase" serialized:"phase"`
}

func (c *command) newComputePoolCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "compute-pool",
		Short:       "Manage Flink compute pools in Confluent Platform.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.AddCommand(c.newComputePoolCreateCommandOnPrem())
	cmd.AddCommand(c.newComputePoolDeleteCommandOnPrem())
	cmd.AddCommand(c.newComputePoolDescribeCommandOnPrem())
	cmd.AddCommand(c.newComputePoolListCommandOnPrem())

	return cmd
}

func convertSdkComputePoolToLocalComputePool(sdkComputePool cmfsdk.ComputePool) LocalComputePool {
	localPool := LocalComputePool{
		ApiVersion: sdkComputePool.ApiVersion,
		Kind:       sdkComputePool.Kind,
		Metadata: LocalComputePoolMetadata{
			Name:              sdkComputePool.Metadata.Name,
			CreationTimestamp: sdkComputePool.Metadata.CreationTimestamp,
			Uid:               sdkComputePool.Metadata.Uid,
			Labels:            sdkComputePool.Metadata.Labels,
			Annotations:       sdkComputePool.Metadata.Annotations,
		},
		Spec: LocalComputePoolSpec{
			Type:        sdkComputePool.Spec.Type,
			ClusterSpec: sdkComputePool.Spec.ClusterSpec,
		},
	}

	if phase := extractComputePoolPhase(sdkComputePool); phase != "" {
		localPool.Status = &LocalComputePoolStatus{
			Phase: phase,
		}
	}

	return localPool
}

func extractComputePoolPhase(pool cmfsdk.ComputePool) string {
	phase, _ := flink.GetMapField[string](pool.GetStatus(), "phase", fmt.Sprintf("compute pool %q", pool.GetMetadata().Name))
	return phase
}
