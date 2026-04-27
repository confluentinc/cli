//go:build live_test && (all || rtce)

package live

import (
	"strings"
	"testing"
	"time"
)

func (s *CLILiveTestSuite) TestRtceTopicCRUDLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	// Variables
	rtceTopicName := uniqueName("rtceto")
	description := "Live test description"
	updatedDescription := "Updated live test description"

	// Cleanup (LIFO)
	s.registerCleanup(t, "environment delete {{.env_id}} --force", state)
	s.registerCleanup(t, "rtce rtce-topic delete {{.rtce_topic_id}} --environment {{.env_id}} --force", state)

	// Phase 1: Create steps
	createSteps := []CLILiveTest{
		{
			Name:            "Create environment",
			Args:            "environment create " + uniqueName("env") + " -o json",
			JSONFieldsExist: []string{"id"},
			CaptureID:       "env_id",
		},
		{
			Name:         "Use environment",
			Args:         "environment use {{.env_id}}",
			UseStateVars: true,
		},
		{
			Name:            "Create rtce topic",
			Args:            "rtce rtce-topic create " + rtceTopicName + " --description \"" + description + "\" --region \"" + region + "\" --topic-name \"" + topicName + "\" --environment {{.env_id}} -o json",
			UseStateVars:    true,
			CaptureID:       "rtce_topic_id",
			JSONFields:      map[string]string{},
			JSONFieldsExist: []string{"id"},
		},
	}

	for _, step := range createSteps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}

	// Phase 2: Wait for provisioning
	t.Run("Wait for rtce topic provisioned", func(t *testing.T) {
		s.waitForCondition(t,
			"rtce rtce-topic describe {{.rtce_topic_id}} --environment {{.env_id}} -o json",
			state,
			func(output string) bool {
				status := extractJSONField(t, output, "phase")
				return strings.EqualFold(status, "ACTIVE")
			},
			30*time.Second,
			10*time.Minute,
		)
	})

	// Phase 3: CRUD operations
	crudSteps := []CLILiveTest{
		{
			Name:         "Describe rtce topic",
			Args:         "rtce rtce-topic describe {{.rtce_topic_id}} --environment {{.env_id}} -o json",
			UseStateVars: true,
			JSONFields:   map[string]string{},
		},
		{
			Name:         "List rtce topics",
			Args:         "rtce rtce-topic list --environment {{.env_id}}",
			UseStateVars: true,
			Contains:     []string{rtceTopicName},
		},
		{
			Name:         "Update rtce topic description",
			Args:         "rtce rtce-topic update {{.rtce_topic_id}} --description \"" + updatedDescription + "\" --environment {{.env_id}}",
			UseStateVars: true,
		},
		{
			Name:         "Describe updated rtce topic",
			Args:         "rtce rtce-topic describe {{.rtce_topic_id}} --environment {{.env_id}} -o json",
			UseStateVars: true,
			JSONFields: map[string]string{
				"description": updatedDescription,
			},
		},
		{
			Name:         "Delete rtce topic",
			Args:         "rtce rtce-topic delete {{.rtce_topic_id}} --environment {{.env_id}} --force",
			UseStateVars: true,
		},
		{
			Name:         "Verify deletion",
			Args:         "rtce rtce-topic describe {{.rtce_topic_id}} --environment {{.env_id}}",
			UseStateVars: true,
			ExitCode:     1,
		},
	}

	for _, step := range crudSteps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
