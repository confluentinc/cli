package cmd

import (
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

func IsAuditLogsEnabled(state *v1.ContextState) (*orgv1.AuditLog, bool) {
	if state.Auth == nil || state.Auth.Organization == nil || state.Auth.Organization.AuditLog == nil {
		return nil, false
	}
	return state.Auth.Organization.AuditLog, true
}
