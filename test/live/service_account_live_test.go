//go:build live_test && (all || core)

package live

import (
	"testing"
)

func (s *CLILiveTestSuite) TestServiceAccountCRUDLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	saName := uniqueName("sa")
	saDescription := "Live test service account"
	updatedDescription := "Updated live test service account"

	// Register cleanup immediately
	s.registerCleanup(t, "iam service-account delete {{.sa_id}} --force", state)

	steps := []CLILiveTest{
		{
			Name:     "Create service account",
			Args:     `iam service-account create ` + saName + ` --description "` + saDescription + `" -o json`,
			ExitCode: 0,
			JSONFields: map[string]string{
				"name":        saName,
				"description": saDescription,
			},
			JSONFieldsExist: []string{"id"},
			CaptureID:       "sa_id",
		},
		{
			Name:         "Describe service account",
			Args:         "iam service-account describe {{.sa_id}} -o json",
			UseStateVars: true,
			ExitCode:     0,
			JSONFields: map[string]string{
				"name": saName,
			},
		},
		{
			Name:     "List service accounts",
			Args:     "iam service-account list",
			ExitCode: 0,
			Contains: []string{saName},
		},
		{
			Name:         "Update service account description",
			Args:         `iam service-account update {{.sa_id}} --description "` + updatedDescription + `"`,
			UseStateVars: true,
			ExitCode:     0,
		},
		{
			Name:         "Describe updated service account",
			Args:         "iam service-account describe {{.sa_id}} -o json",
			UseStateVars: true,
			ExitCode:     0,
			JSONFields: map[string]string{
				"description": updatedDescription,
			},
		},
		{
			Name:         "Delete service account",
			Args:         "iam service-account delete {{.sa_id}} --force",
			UseStateVars: true,
			ExitCode:     0,
		},
		{
			Name:         "Verify deletion",
			Args:         "iam service-account describe {{.sa_id}}",
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
