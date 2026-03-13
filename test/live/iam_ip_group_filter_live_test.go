//go:build live_test && (all || iam)

package live

import (
	"testing"
)

func (s *CLILiveTestSuite) TestIAMIpGroupFilterCRUDLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	ipGroupName := uniqueName("ipgrp")
	updatedIpGroupName := ipGroupName + "-updated"
	ipFilterName := uniqueName("ipflt")

	// Register cleanups in LIFO order: delete filter first, then group
	s.registerCleanup(t, "iam ip-group delete {{.ip_group_id}} --force", state)
	s.registerCleanup(t, "iam ip-filter delete {{.ip_filter_id}} --force", state)

	steps := []CLILiveTest{
		// IP Group CRUD
		{
			Name:      "Create IP group",
			Args:      "iam ip-group create " + ipGroupName + " --cidr-blocks 10.0.0.0/24 -o json",
			CaptureID: "ip_group_id",
			JSONFields: map[string]string{
				"group_name": ipGroupName,
			},
		},
		{
			Name:         "Describe IP group",
			Args:         "iam ip-group describe {{.ip_group_id}} -o json",
			UseStateVars: true,
			JSONFields: map[string]string{
				"group_name": ipGroupName,
			},
		},
		{
			Name:     "List IP groups",
			Args:     "iam ip-group list",
			Contains: []string{ipGroupName},
		},
		{
			Name:         "Update IP group",
			Args:         "iam ip-group update {{.ip_group_id}} --name " + updatedIpGroupName + " --cidr-blocks 10.0.0.0/24,10.1.0.0/24",
			UseStateVars: true,
		},
		{
			Name:         "Verify IP group updated",
			Args:         "iam ip-group describe {{.ip_group_id}} -o json",
			UseStateVars: true,
			JSONFields: map[string]string{
				"group_name": updatedIpGroupName,
			},
		},
		// IP Filter CRUD (depends on the IP group)
		{
			Name:         "Create IP filter",
			Args:         "iam ip-filter create " + ipFilterName + " --ip-groups {{.ip_group_id}} --operations MANAGEMENT -o json",
			UseStateVars: true,
			CaptureID:    "ip_filter_id",
			JSONFields: map[string]string{
				"filter_name": ipFilterName,
			},
		},
		{
			Name:         "Describe IP filter",
			Args:         "iam ip-filter describe {{.ip_filter_id}} -o json",
			UseStateVars: true,
			JSONFields: map[string]string{
				"filter_name": ipFilterName,
			},
		},
		{
			Name:     "List IP filters",
			Args:     "iam ip-filter list",
			Contains: []string{ipFilterName},
		},
		{
			Name:         "Delete IP filter",
			Args:         "iam ip-filter delete {{.ip_filter_id}} --force",
			UseStateVars: true,
		},
		// Clean up IP group
		{
			Name:         "Delete IP group",
			Args:         "iam ip-group delete {{.ip_group_id}} --force",
			UseStateVars: true,
		},
		{
			Name:         "Verify IP group deleted",
			Args:         "iam ip-group describe {{.ip_group_id}}",
			UseStateVars: true,
			ExitCode:     1,
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
