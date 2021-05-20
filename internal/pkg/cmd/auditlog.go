package cmd

import (
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"

	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
)

func IsAuditLogsEnabled(state *v2.ContextState) (*orgv1.AuditLog, bool) {
	if state.Auth == nil || state.Auth.Organization == nil || state.Auth.Organization.AuditLog == nil {
		return nil, false
	}
	return state.Auth.Organization.AuditLog, true
}
