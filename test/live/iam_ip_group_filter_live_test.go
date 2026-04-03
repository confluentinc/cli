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

	// Register cleanup
	s.registerCleanup(t, "iam ip-group delete {{.ip_group_id}} --force", state)

	steps := []CLILiveTest{
		// IP Group CRUD
		{
			Name:      "Create IP group",
			Args:      "iam ip-group create " + ipGroupName + " --cidr-blocks 10.0.0.0/24 -o json",
			CaptureID: "ip_group_id",
			JSONFields: map[string]string{
				"name": ipGroupName,
			},
		},
		{
			Name:         "Describe IP group",
			Args:         "iam ip-group describe {{.ip_group_id}} -o json",
			UseStateVars: true,
			JSONFields: map[string]string{
				"name": ipGroupName,
			},
		},
		{
			Name:     "List IP groups",
			Args:     "iam ip-group list",
			Contains: []string{ipGroupName},
		},
		{
			Name:         "Update IP group",
			Args:         "iam ip-group update {{.ip_group_id}} --name " + updatedIpGroupName + " --add-cidr-blocks 10.1.0.0/24",
			UseStateVars: true,
		},
		{
			Name:         "Verify IP group updated",
			Args:         "iam ip-group describe {{.ip_group_id}} -o json",
			UseStateVars: true,
			JSONFields: map[string]string{
				"name": updatedIpGroupName,
			},
		},
		// Note: IP Filter CRUD steps are skipped because creating a MANAGEMENT filter
		// without including the caller's IP triggers a lockout protection error.
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
