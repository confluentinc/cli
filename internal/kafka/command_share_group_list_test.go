package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

func TestShareGroupListCommand(t *testing.T) {
	cfg := &config.Config{}
	prerunner := &pcmd.PreRun{Config: cfg}

	// Test that the command can be created
	cmd := newShareCommand(cfg, prerunner)
	require.NotNil(t, cmd)
	require.Equal(t, "share", cmd.Use)
	require.Equal(t, "Manage Kafka shares.", cmd.Short)

	// Test that the group subcommand exists
	groupCmd := cmd.Commands()[0]
	require.NotNil(t, groupCmd)
	require.Equal(t, "group", groupCmd.Use)
	require.Equal(t, "Manage Kafka share groups.", groupCmd.Short)

	// Test that the list subcommand exists
	listCmd := groupCmd.Commands()[0]
	require.NotNil(t, listCmd)
	require.Equal(t, "list", listCmd.Use)
	require.Equal(t, "List Kafka share groups.", listCmd.Short)

	// Test that the list command has the expected flags
	require.NotNil(t, listCmd.Flags().Lookup("cluster"))
	require.NotNil(t, listCmd.Flags().Lookup("output"))
}

func TestShareGroupOutStruct(t *testing.T) {
	// Test that the struct has the expected fields
	group := &shareGroupOut{
		Cluster:        "cluster-1",
		ShareGroup:     "share-group-1",
		Coordinator:    "1",
		State:          "STABLE",
		ConsumerCount:  2,
		PartitionCount: 3,
	}

	require.Equal(t, "cluster-1", group.Cluster)
	require.Equal(t, "share-group-1", group.ShareGroup)
	require.Equal(t, "1", group.Coordinator)
	require.Equal(t, "STABLE", group.State)
	require.Equal(t, int32(2), group.ConsumerCount)
	require.Equal(t, int32(3), group.PartitionCount)
}
