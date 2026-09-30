//go:build live_test && (all || iam)

package live

import (
	"os"
	"testing"
)

func (s *CLILiveTestSuite) TestRBACRoleBindingCRUDLive() {
	t := s.T()
	t.Parallel()

	envID := os.Getenv("LIVE_TEST_ENVIRONMENT_ID")
	if envID == "" {
		t.Skip("Skipping: LIVE_TEST_ENVIRONMENT_ID must be set")
	}

	state := s.setupTestContext(t)

	saName := uniqueName("rbac-sa")

	// Cleanup in LIFO order: delete role binding, then service account
	s.registerCleanup(t, "iam service-account delete {{.sa_id}} --force", state)
	s.registerCleanup(t, "iam rbac role-binding delete --principal User:{{.sa_id}} --role EnvironmentAdmin --environment "+envID+" --force", state)

	steps := []CLILiveTest{
		{
			Name: "List available roles",
			Args: "iam rbac role list",
		},
		{
			Name:            "Create service account for RBAC test",
			Args:            `iam service-account create ` + saName + ` --description "SA for RBAC live test" -o json`,
			JSONFieldsExist: []string{"id"},
			CaptureID:       "sa_id",
		},
		{
			Name:         "Create role binding (EnvironmentAdmin)",
			Args:         "iam rbac role-binding create --principal User:{{.sa_id}} --role EnvironmentAdmin --environment " + envID,
			UseStateVars: true,
		},
		{
			Name:         "List role bindings for principal",
			Args:         "iam rbac role-binding list --principal User:{{.sa_id}} --environment " + envID,
			UseStateVars: true,
			Contains:     []string{"EnvironmentAdmin"},
			Retries:      5,
		},
		{
			Name:         "Delete role binding",
			Args:         "iam rbac role-binding delete --principal User:{{.sa_id}} --role EnvironmentAdmin --environment " + envID + " --force",
			UseStateVars: true,
		},
		{
			Name:         "Verify role binding deleted",
			Args:         "iam rbac role-binding list --principal User:{{.sa_id}} --environment " + envID,
			UseStateVars: true,
			NotContains:  []string{"EnvironmentAdmin"},
			Retries:      5,
		},
		{
			Name:         "Delete service account",
			Args:         "iam service-account delete {{.sa_id}} --force",
			UseStateVars: true,
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
