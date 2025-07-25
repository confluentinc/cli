package flink

import (

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newComputePoolListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List Flink compute pools in Confluent Platform.",
		Args:        cobra.NoArgs,
		RunE:        c.computePoolListOnPrem,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) computePoolListOnPrem(cmd *cobra.Command, _ []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkComputePools, err := client.ListComputePools(c.createContext(), environment)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, pool := range sdkComputePools {
			var creationTime string
			if pool.GetMetadata().CreationTimestamp != nil {
				creationTime = *pool.GetMetadata().CreationTimestamp
			} else {
				creationTime = ""
			}
			list.Add(&computePoolOutOnPrem{
				CreationTime: creationTime,
				Name:         pool.GetMetadata().Name,
				Type:         pool.GetSpec().Type,
				Phase:        pool.GetStatus().Phase,
			})
		}
		return list.Print()
	}

	// Create the slice to hold the clean objects
localPools := make([]LocalComputePool, 0, len(sdkComputePools))

// Loop through the original SDK objects
for _, sdkPool := range sdkComputePools {

	// --- Start Deep Copy for each item ---

	// Build the full LocalComputePool with all nested Local* types
	localPool := LocalComputePool{
		ApiVersion: sdkPool.ApiVersion,
		Kind:       sdkPool.Kind,
		Metadata: LocalComputePoolMetadata{
			Name:              sdkPool.Metadata.Name,
			CreationTimestamp: sdkPool.Metadata.CreationTimestamp,
			Uid:               sdkPool.Metadata.Uid,
			Labels:            sdkPool.Metadata.Labels,
			Annotations:       sdkPool.Metadata.Annotations,
		},
		Spec: LocalComputePoolSpec{
			Type:        sdkPool.Spec.Type,
			ClusterSpec: sdkPool.Spec.ClusterSpec,
		},
	}

	// Handle the status, which is a pointer and could be nil.
	if sdkPool.Status != nil {
		localPool.Status = &LocalComputePoolStatus{
			Phase: sdkPool.Status.Phase,
		}
	}

	// Append the fully "clean" object to the final slice
	localPools = append(localPools, localPool)
}

// Now, localPools is ready to be serialized
return output.SerializedOutput(cmd, localPools)
}
