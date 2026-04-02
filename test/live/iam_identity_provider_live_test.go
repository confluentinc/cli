//go:build live_test && (all || iam)

package live

import (
	"testing"
)

func (s *CLILiveTestSuite) TestIAMIdentityProviderCRUDLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	providerName := uniqueName("idp")
	updatedProviderName := providerName + "-updated"
	poolName := uniqueName("pool")
	groupMappingName := uniqueName("grpmap")

	// Register cleanups in LIFO order: group-mapping, pool, then provider
	s.registerCleanup(t, "iam provider delete {{.provider_id}} --force", state)
	s.registerCleanup(t, "iam pool delete {{.pool_id}} --provider {{.provider_id}} --force", state)
	s.registerCleanup(t, "iam group-mapping delete {{.group_mapping_id}} --force", state)

	steps := []CLILiveTest{
		// Identity Provider CRUD
		{
			Name:      "Create identity provider",
			Args:      "iam provider create " + providerName + ` --description "Live test provider" --issuer-uri https://example.com/issuer --jwks-uri https://example.com/.well-known/jwks.json -o json`,
			CaptureID: "provider_id",
			JSONFields: map[string]string{
				"name": providerName,
			},
		},
		{
			Name:         "Describe identity provider",
			Args:         "iam provider describe {{.provider_id}} -o json",
			UseStateVars: true,
			JSONFields: map[string]string{
				"name": providerName,
			},
		},
		{
			Name:     "List identity providers",
			Args:     "iam provider list",
			Contains: []string{providerName},
		},
		{
			Name:         "Update identity provider",
			Args:         "iam provider update {{.provider_id}} --name " + updatedProviderName + ` --description "Updated live test provider"`,
			UseStateVars: true,
		},
		{
			Name:         "Verify provider updated",
			Args:         "iam provider describe {{.provider_id}} -o json",
			UseStateVars: true,
			JSONFields: map[string]string{
				"name": updatedProviderName,
			},
		},
		// Identity Pool CRUD (depends on provider)
		{
			Name:         "Create identity pool",
			Args:         `iam pool create ` + poolName + ` --provider {{.provider_id}} --identity-claim sub --description "Live test pool" -o json`,
			UseStateVars: true,
			CaptureID:    "pool_id",
			JSONFields: map[string]string{
				"display_name": poolName,
			},
		},
		{
			Name:         "Describe identity pool",
			Args:         "iam pool describe {{.pool_id}} --provider {{.provider_id}} -o json",
			UseStateVars: true,
			JSONFields: map[string]string{
				"display_name": poolName,
			},
		},
		{
			Name:         "List identity pools",
			Args:         "iam pool list --provider {{.provider_id}}",
			UseStateVars: true,
			Contains:     []string{poolName},
		},
		{
			Name:         "Update identity pool",
			Args:         `iam pool update {{.pool_id}} --provider {{.provider_id}} --description "Updated live test pool"`,
			UseStateVars: true,
		},
		{
			Name:         "Delete identity pool",
			Args:         "iam pool delete {{.pool_id}} --provider {{.provider_id}} --force",
			UseStateVars: true,
		},
		// Group Mapping CRUD
		{
			Name:      "Create group mapping",
			Args:      `iam group-mapping create ` + groupMappingName + ` --description "Live test group mapping" --filter '"engineering" in groups' -o json`,
			CaptureID: "group_mapping_id",
			JSONFields: map[string]string{
				"display_name": groupMappingName,
			},
		},
		{
			Name:         "Describe group mapping",
			Args:         "iam group-mapping describe {{.group_mapping_id}} -o json",
			UseStateVars: true,
			JSONFields: map[string]string{
				"display_name": groupMappingName,
			},
		},
		{
			Name:     "List group mappings",
			Args:     "iam group-mapping list",
			Contains: []string{groupMappingName},
		},
		{
			Name:         "Update group mapping",
			Args:         `iam group-mapping update {{.group_mapping_id}} --description "Updated live test group mapping"`,
			UseStateVars: true,
		},
		{
			Name:         "Delete group mapping",
			Args:         "iam group-mapping delete {{.group_mapping_id}} --force",
			UseStateVars: true,
		},
		// Clean up provider (after pool is deleted)
		{
			Name:         "Delete identity provider",
			Args:         "iam provider delete {{.provider_id}} --force",
			UseStateVars: true,
		},
		{
			Name:         "Verify provider deleted",
			Args:         "iam provider describe {{.provider_id}}",
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
